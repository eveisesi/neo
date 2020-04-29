package esi

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/eveisesi/neo"
	"github.com/volatiletech/null"

	"github.com/go-redis/redis"
	"github.com/pkg/errors"
)

var (
	err error
	mx  sync.Mutex
)

type (
	Service interface {
		GetAlliancesAllianceID(id uint64, etag null.String) (*neo.Alliance, *Meta)
		GetCharactersCharacterID(id uint64, etag null.String) (*neo.Character, *Meta)
		GetCorporationsCorporationID(id uint64, etag null.String) (*neo.Corporation, *Meta)
		GetKillmailsKillmailIDKillmailHash(id, hash string) (*neo.Killmail, *Meta)
		HeadMarketsRegionIDTypes(regionID uint64) *Meta
		GetMarketGroups() ([]int, *Meta)
		GetMarketGroupsMarketGroupID(id int) (*neo.MarketGroup, *Meta)
		GetMarketsRegionIDTypes(regionID uint64, page null.String) ([]int, *Meta)
		GetMarketsRegionIDHistory(regionID uint64, typeID string) ([]*neo.HistoricalRecord, *Meta)
		GetUniverseSystemsSystemID(id uint64) (*neo.SolarSystem, *Meta)
		GetUniverseTypesTypeID(id uint64) (*neo.Type, []*neo.TypeAttribute, *Meta)
	}
	service struct {
		client *http.Client
		redis  *redis.Client
		ua     string
		// remain        uint64 // Number of Error left until a 420 will be thrown
		// reset         uint64 // Number of Seconds remain until Remain is reset to 100
		maxattempts uint64
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
	}
)

func newMeta(method, path, query string, code int, headers map[string]string, msg error) *Meta {
	return &Meta{method, path, query, code, headers, msg}
}

func (r *Meta) Error() string {
	return r.Msg.Error()
}

func (r *Meta) IsError() bool {
	return r.Msg != nil
}

// New returns a default configuration for this package
func New(redis *redis.Client, host, uagent string) Service {

	return &service{
		redis: redis,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		ua:          uagent,
		maxattempts: 3,
	}

}

// Request prepares and executes an http request to the EVE Swagger Interface OpenAPI
// and returns the response
func (s *service) request(r request) ([]byte, *Meta) {

	count, err := s.redis.Get(neo.REDIS_ESI_ERROR_COUNT).Int64()
	if err != nil && err.Error() != neo.ErrRedisNil.Error() {
		return nil, newMeta(r.method, r.path, r.query, 500, map[string]string{}, errors.Wrap(err, "unable to get error count from ESI"))
	}

	if count < 10 {
		return nil, newMeta(r.method, r.path, r.query, 500, map[string]string{}, errors.Wrap(err, "420 prone, backing off"))
	}

	status, err := s.redis.Get(neo.REDIS_ESI_ALERT_STATUS).Int64()
	if err != nil && err.Error() != neo.ErrRedisNil.Error() {
		return nil, newMeta(r.method, r.path, r.query, 500, map[string]string{}, errors.Wrap(err, "unable to get 420 status from Redis"))
	}

	switch status {
	case 2:
		time.Sleep(time.Millisecond * 100)
		return nil, newMeta(r.method, r.path, r.query, 500, map[string]string{}, errors.New("currently red alert, unable to make request"))
	case 1:
		time.Sleep(time.Millisecond * 250)
	default:
		break
	}

	uri := url.URL{
		Scheme:   "https",
		Host:     "esi.evetech.net",
		Path:     r.path,
		RawQuery: r.query,
	}

	req, err := http.NewRequest(r.method, uri.String(), bytes.NewBuffer(r.body))
	if err != nil {
		err = errors.Wrap(err, "Unable build request")
		return nil, newMeta(r.method, r.path, r.query, -1, map[string]string{}, errors.Wrap(err, "failed to build esi request"))
	}

	for k, v := range r.headers {
		req.Header.Add(k, v)
	}

	req.Header.Add("Content-Type", "application/json; charset=UTF-8")
	req.Header.Add("User-Agent", s.ua)

	m := newMeta(r.method, r.path, r.query, -1, map[string]string{}, nil)

	attempts := uint64(0)
	var httpResponse *http.Response

	for {

		if attempts >= s.maxattempts {
			m.Msg = errors.New("max attempts exceeded")
			return nil, m
		}

		httpResponse, err = s.client.Do(req)
		if err != nil {
			err = errors.Wrap(err, "failed to make esi request")
			return nil, newMeta(r.method, r.path, r.query, -1, map[string]string{}, err)
		}

		if httpResponse.StatusCode < 400 {
			break
		}

		attempts++
		time.Sleep(time.Second * 2)

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
		return nil, newMeta(r.method, r.path, r.query, httpResponse.StatusCode, headers, errors.Wrap(err, "failed to build esi request"))
	}

	httpResponse.Body.Close()

	m = newMeta(r.method, r.path, r.query, httpResponse.StatusCode, headers, nil)

	_ = s.retrieveErrorReset(headers)
	_ = s.retrieveErrorCount(headers)

	if m.Code < 400 {
		_ = s.reportSuccessfullyESICall()
	}

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
// has triggered and how many more we can trigger before being 420'd
func (s *service) retrieveErrorCount(h map[string]string) error {
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
	_, err = s.redis.Set(neo.REDIS_ESI_ERROR_COUNT, count, 0).Result()
	mx.Unlock()

	return err
}

// retrieveErrorReset is a helper method that retrieves the number of seconds until our Error Limit resets
func (s *service) retrieveErrorReset(h map[string]string) error {
	if _, ok := h["X-Esi-Error-Limit-Reset"]; !ok {
		err = fmt.Errorf("X-Esi-Error-Limit-Reset Header is missing")
		return err
	}

	seconds, err := strconv.ParseUint(h["X-Esi-Error-Limit-Reset"], 10, 32)
	if err != nil {
		return err
	}

	mx.Lock()
	_, err = s.redis.Set(neo.REDIS_ESI_ERROR_RESET, seconds, 0).Result()
	mx.Unlock()

	return err
}

func (s *service) reportSuccessfullyESICall() error {

	value := time.Now().UnixNano()
	s.redis.ZAdd("esi:tracking:success", redis.Z{Score: float64(value), Member: strconv.FormatInt(value, 10)})

	return err

}
