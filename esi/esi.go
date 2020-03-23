package esi

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/pkg/errors"
)

var (
	layoutESI = "Mon, 02 Jan 2006 15:04:05 MST"
	err       error
	mx        sync.Mutex
)

type (
	// Client represents the application as a whole. Client has our HTTP Client, DB Client, and holds Secrets for Third Party API Communication

	Client struct {
		Host          string
		Http          *http.Client
		UserAgent     string
		Remain        uint64 // Number of Error left until a 420 will be thrown
		Reset         uint64 // Number of Seconds remain until Remain is reset to 100
		MaxAttempts   uint64
		SleepDuration time.Duration
	}
	Config struct {
		Host      string `envconfig:"ESI_HOST" required:"true"`
		UserAgent string `envconfig:"API_USER_AGENT" required:"true"`
	}

	Request struct {
		Method  string
		Path    url.URL
		Headers map[string]string
		Body    []byte
	}

	Response struct {
		Method  string
		Path    string
		Code    int
		Headers map[string]string
		Data    interface{}
	}
)

// New returns a default configuration for this package
func New(client *http.Client, host, uagent string) *Client {

	if client == nil {
		client = &http.Client{
			Timeout: 30 * time.Second,
		}
	}

	return &Client{
		Host:          host,
		Http:          client,
		UserAgent:     uagent,
		Remain:        100,
		Reset:         60,
		MaxAttempts:   3,
		SleepDuration: time.Second * 2,
	}

}

// Request prepares and executes an http request to the EVE Swagger Interface OpenAPI
// and returns the response
func (e *Client) Request(request Request) (Response, error) {

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
	req.Header.Add("User-Agent", e.UserAgent)

	resp, err := e.Http.Do(req)
	if err != nil {
		err = errors.Wrap(err, "Unable to make request")
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

	mx.Lock()
	e.Reset = RetrieveErrorResetFromResponse(response)
	e.Remain = RetrieveErrorCountFromResponse(response)
	mx.Unlock()

	return response, nil
}

// RetrieveExpiresHeaderFromResponse takes a response and pull Expires header from the headers. If duraction
// is greater than zero(0), then that number of minutes will be add to the expires time that is parsed from the header.
func RetrieveExpiresHeaderFromResponse(response Response, duration int) (time.Time, error) {
	if _, ok := response.Headers["Expires"]; !ok {
		err := fmt.Errorf("Expires Headers is missing for url %s", response.Path)
		return time.Time{}, err
	}
	expires, err := time.Parse(layoutESI, response.Headers["Expires"])
	if err != nil {
		return expires, err
	}

	if duration > 0 {
		expires = expires.Add(time.Minute * time.Duration(duration))
	}

	return expires, nil
}

// RetrieveEtagHeaderFromResponse is a helper method that retrieves an Etag for the most recent request to
// ESI
func RetrieveEtagHeaderFromResponse(response Response) (string, error) {
	if _, ok := response.Headers["Etag"]; !ok {
		err = fmt.Errorf("Etag Header is missing from url %s", response.Path)
		return "", err
	}
	return response.Headers["Etag"], nil
}

// RetrieveErrorCountFromResponse is a helper method that retrieves the number of errors that this application
// has triggered and how many more we can trigger before being 420'd
func RetrieveErrorCountFromResponse(response Response) uint64 {
	if _, ok := response.Headers["X-Esi-Error-Limit-Remain"]; !ok {
		err = fmt.Errorf("X-Esi-Error-Limit-Remain Header is missing from url %s", response.Path)
		return 100
	}

	count, err := strconv.ParseUint(response.Headers["X-Esi-Error-Limit-Remain"], 10, 32)
	if err != nil {
		return 100
	}

	return count
}

// RetrieveErrorResetFromResponse is a helper method that retrieves the number of seconds until our Error Limit resets
func RetrieveErrorResetFromResponse(response Response) uint64 {
	if _, ok := response.Headers["X-Esi-Error-Limit-Reset"]; !ok {
		err = fmt.Errorf("X-Esi-Error-Limit-Reset Header is missing from url %s", response.Path)
		return 100
	}
	seconds, err := strconv.ParseUint(response.Headers["X-Esi-Error-Limit-Reset"], 10, 32)
	if err != nil {
		return 100
	}

	return seconds
}
