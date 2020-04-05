package token

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/eveisesi/neo"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/pkg/errors"
)

func (s *service) GetState(state string) string {
	return s.oauth.AuthCodeURL(state)
}

func (s *service) GetTokenForCode(ctx context.Context, state, code string) (*neo.Token, error) {

	// Exchange code for token from Oauth2.0 Service
	token, err := s.oauth.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}

	parser := new(jwt.Parser)
	parser.UseJSONNumber = true

	parsed, err := parser.Parse(token.AccessToken, s.getSignatureKey)
	if err != nil {
		s.logger.WithError(err).Error("unable to parse token")
		return nil, errors.Wrap(err, "failed to parse JWT Token")
	}

	characterID, err := strconv.ParseUint(strings.Split(parsed.Claims.(jwt.MapClaims)["sub"].(string), ":")[2], 10, 64)
	if err != nil {
		return nil, errors.Wrap(err, "unable to coerce string to int")
	}

	// // Check to see if we know who this character is
	neoToken, err := s.Token(ctx, characterID)
	if err != nil && err != sql.ErrNoRows {
		return nil, errors.Wrap(err, "unexpected error encountered")
	}

	if err == sql.ErrNoRows {

		neoToken = &neo.Token{
			ID:           characterID,
			AccessToken:  token.AccessToken,
			RefreshToken: token.RefreshToken,
			Expiry:       time.Now().Add(time.Minute * 19),
		}

		neoToken, err = s.CreateToken(ctx, neoToken)
		if err != nil {
			return nil, errors.Wrap(err, "unable to create token")
		}
	}
	if err == nil && neoToken != nil {
		neoToken = &neo.Token{
			AccessToken:  token.AccessToken,
			RefreshToken: token.RefreshToken,
			Expiry:       time.Now().Add(time.Minute * 19),
		}

		neoToken, err = s.UpdateToken(ctx, characterID, neoToken)
		if err != nil {
			return nil, errors.Wrap(err, "unable to update token")
		}
	}

	return neoToken, errors.Wrap(err, "unable to generate jwt token. key is not set and/or missing")
}

func (s *service) getSignatureKey(token *jwt.Token) (interface{}, error) {

	key := "neo:jwk"
	result, err := s.redis.Get(key).Bytes()
	if err != nil && err.Error() != "redis: nil" {
		return nil, errors.Wrap(err, "unexpected error looking for jwk in redis")
	}

	if len(result) == 0 {
		res, err := s.client.Get(s.jwksURL)
		if err != nil {
			return nil, errors.Wrap(err, "unable to retrieve jwks from sso")
		}

		if res.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("unexpected status code recieved while fetching jwks. %d", res.StatusCode)
		}

		buf, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, errors.Wrap(err, "faile dto read jwk response body")
		}

		_, err = s.redis.Set(key, buf, time.Minute*3600).Result()
		if err != nil {
			return nil, errors.Wrap(err, "failed to cache jwks in redis")
		}

		result = buf

	}

	set, err := jwk.ParseBytes(result)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse jwks bytes")
	}

	keyID, ok := token.Header["kid"].(string)
	if !ok {
		return nil, errors.New("expected jwt header to have string kid")
	}

	webkey := set.LookupKeyID(keyID)
	if len(webkey) == 1 {
		return webkey[0].Materialize()
	}

	return nil, fmt.Errorf("unable to find key with kid of %s", keyID)
}

func (s *service) request(request Request) (Response, error) {

	var rBody io.Reader

	if request.Body != nil {
		rBody = bytes.NewBuffer(request.Body)
	}

	req, err := http.NewRequest(request.Method, request.Path.String(), rBody)
	if err != nil {
		err = errors.Wrap(err, "Unable build request")
		return Response{}, err
	}
	for k, v := range request.Headers {
		req.Header.Add(k, v)
	}

	req.Header.Add("Content-Type", "application/json; charset=UTF-8")
	// req.Header.Add("User-Agent", e.UserAgent)

	resp, err := s.client.Do(req)
	if err != nil {
		return Response{}, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err = errors.Wrap(err, "error reading body")
		return Response{}, err
	}

	resp.Body.Close()

	var response Response
	response.Method = request.Method
	response.Path = request.Path.Path
	response.Data = body
	response.Code = resp.StatusCode
	headers := make(map[string]string)
	for k, sv := range resp.Header {
		for _, v := range sv {
			headers[k] = v
		}
	}

	response.Headers = headers

	return response, nil
}
