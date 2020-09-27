package esi

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"runtime/debug"
	"strconv"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/eveisesi/neo"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/volatiletech/null"

	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
)

var (
	err error
	mx  sync.Mutex
)

type (
	Service interface {
		// Alliances
		GetAlliancesAllianceID(ctx context.Context, id uint, etag string) (*neo.Alliance, Meta)

		// Characters
		GetCharactersCharacterID(ctx context.Context, id uint64, etag string) (*neo.Character, Meta)

		// Corporations
		GetCorporationsCorporationID(ctx context.Context, id uint, etag string) (*neo.Corporation, Meta)

		// Killmails
		GetKillmailsKillmailIDKillmailHash(ctx context.Context, id uint, hash string) (*neo.Killmail, Meta)

		// Market
		HeadMarketsRegionIDTypes(ctx context.Context, regionID uint) Meta
		GetMarketGroups(ctx context.Context) ([]int, Meta)
		GetMarketGroupsMarketGroupID(ctx context.Context, id int) (*neo.MarketGroup, Meta)
		GetMarketsRegionIDTypes(ctx context.Context, regionID uint, page null.String) ([]int, Meta)
		GetMarketsRegionIDHistory(ctx context.Context, regionID uint, typeID uint) ([]*neo.HistoricalRecord, Meta)
		GetMarketsPrices(ctx context.Context) ([]*neo.MarketPrices, Meta)

		// Status
		GetStatus(ctx context.Context) (*neo.ServerStatus, Meta)

		// Universe
		GetUniverseSystemsSystemID(ctx context.Context, id uint) (*neo.SolarSystem, Meta)
		GetUniverseTypesTypeID(ctx context.Context, id uint) (*neo.Type, []*neo.TypeAttribute, Meta)
	}
	service struct {
		client      *http.Client
		redis       *redis.Client
		ua          string
		maxattempts uint
	}

	request struct {
		method  string
		path    string
		query   string
		headers map[string]string
		body    []byte
	}

	Meta struct {
		Method  string
		Path    string
		Query   string
		Code    int
		Headers map[string]string
		Msg     error
		Data    []byte
	}
)

func newMeta(method, path, query string, code int, headers map[string]string, msg error, data []byte) Meta {
	return Meta{method, path, query, code, headers, msg, data}
}

func (r Meta) Error() string {
	if r.Msg == nil {
		return ""
	}
	return r.Msg.Error()
}

func (r Meta) IsErr() bool {
	return r.Msg != nil
}

// New returns a default configuration for this package
func New(redis *redis.Client, host, uagent string) Service {

	client := &http.Client{
		Timeout: time.Second * 3,
	}

	return &service{
		redis:       redis,
		client:      client,
		ua:          uagent,
		maxattempts: 3,
	}

}

// Request prepares and executes an http request to the EVE Swagger Interface OpenAPI
// and returns the response
func (s *service) request(ctx context.Context, r request) ([]byte, Meta) {

	defer func() {
		if recov := recover(); recov != nil {
			spew.Dump(r, recov)
			debug.PrintStack()
		}
	}()

	uri := url.URL{
		Scheme:   "https",
		Host:     "esi.evetech.net",
		Path:     r.path,
		RawQuery: r.query,
	}

	req, err := http.NewRequestWithContext(ctx, r.method, uri.String(), bytes.NewBuffer(r.body))
	if err != nil {
		err = errors.Wrap(err, "Unable build request")
		return nil, newMeta(r.method, r.path, r.query, http.StatusInternalServerError, map[string]string{}, errors.Wrap(err, "failed to build esi request"), nil)
	}
	req = newrelic.RequestWithTransactionContext(req, newrelic.FromContext(ctx))

	for k, v := range r.headers {
		req.Header.Add(k, v)
	}

	req.Header.Add("Content-Type", "application/json; charset=UTF-8")
	req.Header.Add("User-Agent", s.ua)

	m := newMeta(r.method, r.path, r.query, 0, map[string]string{}, nil, []byte{})

	var httpResponse *http.Response
	attempts := uint(0)

	for {

		if attempts >= s.maxattempts {
			if httpResponse != nil && httpResponse.StatusCode > 0 {
				m.Code = httpResponse.StatusCode
			} else {
				m.Code = http.StatusInternalServerError
			}
			m.Msg = errors.New("max attempts exceeded")
			break
		}

		seg := newrelic.StartExternalSegment(newrelic.FromContext(ctx), req)
		httpResponse, err = s.client.Do(req)
		seg.Response = httpResponse
		seg.End()

		if err != nil {
			if _, ok := err.(net.Error); ok {
				attempts++
				time.Sleep(time.Second * 2)
				continue
			}

			err = errors.Wrap(err, "failed to make esi request")

			return nil, newMeta(r.method, r.path, r.query, -1, map[string]string{}, err, []byte{})
		}

		if httpResponse.StatusCode < 500 {
			break
		}

		attempts++
		time.Sleep(time.Second * 2)

	}
	if httpResponse == nil {
		if m.Msg == nil {
			m.Msg = errors.New("failed to successfully make ESI request")
		}
		return nil, m
	}

	headers := make(map[string]string)
	for k, sv := range httpResponse.Header {
		for _, v := range sv {
			headers[k] = v
		}
	}

	data, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		err = errors.Wrap(err, "error reading body")

		return nil, newMeta(r.method, r.path, r.query, httpResponse.StatusCode, headers, errors.Wrap(err, "failed to build esi request"), []byte{})
	}

	httpResponse.Body.Close()

	m = newMeta(r.method, r.path, r.query, httpResponse.StatusCode, headers, nil, data)
	s.trackESICallStatusCode(ctx, m.Code)

	s.retrieveErrorReset(ctx, headers)
	s.retrieveErrorCount(ctx, headers)

	return data, m
}

// retrieveExpiresHeader takes a map[string]string of the response headers, checks to see if the "Expires" key exists, and if it does, parses the timestamp and returns a time.Time. If duraction
// is greater than zero(0), then that number of minutes will be add to the expires time that is parsed from the header.
func (s *service) retrieveExpiresHeader(h map[string]string, duration int) time.Time {
	if _, ok := h["Expires"]; !ok {
		return time.Now().Add(time.Minute * 60)
	}
	expires, err := time.Parse(neo.ESI_EXPIRES_HEADER_FORMAT, h["Expires"])
	if err != nil {
		return expires
	}

	if duration > 0 {
		expires = expires.Add(time.Minute * time.Duration(duration))
	}

	return expires
}

// retrieveEtagHeader is a helper method that retrieves an Etag for the most recent request to
// ESI
func (s *service) retrieveEtagHeader(h map[string]string) string {
	if _, ok := h["Etag"]; !ok {
		return ""
	}
	return h["Etag"]
}

// retrieveErrorCount is a helper method that retrieves the number of errors that this application
// has triggered and how many more can be triggered before potentially encountereding an HTTP Status 420
func (s *service) retrieveErrorCount(ctx context.Context, h map[string]string) {
	// Default to a low count. This will cause the app to slow down
	// if the header is not present to set the actual value from the header
	var count int = 15
	strCount := h["X-Esi-Error-Limit-Remain"]
	if strCount != "" {
		i, err := strconv.Atoi(strCount)
		if err == nil {
			count = i
		}
	}

	mx.Lock()
	s.redis.Set(ctx, neo.REDIS_ESI_ERROR_COUNT, count, 0)
	mx.Unlock()

}

// retrieveErrorReset is a helper method that retrieves the number of seconds until our Error Limit resets
func (s *service) retrieveErrorReset(ctx context.Context, h map[string]string) {
	if _, ok := h["X-Esi-Error-Limit-Reset"]; !ok {
		err = fmt.Errorf("X-Esi-Error-Limit-Reset Header is missing")
		return
	}

	seconds, err := strconv.ParseUint(h["X-Esi-Error-Limit-Reset"], 10, 32)
	if err != nil {
		return
	}

	mx.Lock()
	s.redis.Set(ctx, neo.REDIS_ESI_ERROR_RESET, time.Now().Add(time.Second*time.Duration(seconds)).Unix(), 0)
	mx.Unlock()

}

func (s *service) trackESICallStatusCode(ctx context.Context, code int) {

	value := time.Now().UnixNano()
	input := redis.Z{Score: float64(value), Member: strconv.FormatInt(value, 10)}

	switch n := code; {
	case n == http.StatusOK:
		s.redis.ZAdd(ctx, neo.REDIS_ESI_TRACKING_OK, &input)
	case n == http.StatusNotModified:
		s.redis.ZAdd(ctx, neo.REDIS_ESI_TRACKING_NOT_MODIFIED, &input)
	case n == 420:
		s.redis.ZAdd(ctx, neo.REDIS_ESI_TRACKING_CALM_DOWN, &input)
	case n >= 400 && n < 500:
		s.redis.ZAdd(ctx, neo.REDIS_ESI_TRACKING_4XX, &input)
	case n >= 500:
		s.redis.ZAdd(ctx, neo.REDIS_ESI_TRACKING_5XX, &input)
	}

}
