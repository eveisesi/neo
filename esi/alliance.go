package esi

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/eveisesi/neo"
	"github.com/pkg/errors"
)

// GetAlliancesAllianceID makes a HTTP GET Request to the /alliances/{alliance_id} endpoint
// for information about the provided alliance
//
// Documentation: https://esi.evetech.net/ui/#/Alliance/get_alliances_alliance_id
// Version: v3
// Cache: 3600 sec (1 Hour)
func (e *Client) GetAlliancesAllianceID(id uint64, etag string) (Response, error) {

	var response Response
	path := fmt.Sprintf("/v3/alliances/%d/", id)

	url := url.URL{
		Scheme: "https",
		Host:   e.Host,
		Path:   path,
	}

	headers := make(map[string]string)

	if etag != "" {
		headers["If-None-Match"] = etag
	}

	request := Request{
		Method:  "GET",
		Path:    url,
		Headers: headers,
	}
	attempts := uint64(0)
	for {

		if attempts >= e.MaxAttempts {
			return response, errors.New("max attempts exceeded")
		}

		response, err = e.Request(request)
		if err != nil {
			return response, err
		}
		if response.Code < 400 {
			break
		}

		attempts++
		time.Sleep(time.Second * e.SleepDuration)

	}

	var alliance neo.Alliance

	switch response.Code {
	case 200:
		err = json.Unmarshal(response.Data.([]byte), &alliance)
		if err != nil {
			return response, errors.Wrapf(err, "unable to unmarshel response body on request %s", path)
		}

		alliance.ID = id

		alliance.CachedUntil, err = RetrieveExpiresHeaderFromResponse(response, 0)
		if err != nil {
			return response, errors.Wrapf(err, "Error Encounter with Request %s", path)
		}

		alliance.Etag, err = RetrieveEtagHeaderFromResponse(response)
		if err != nil {
			return response, errors.Wrapf(err, "Error Encounter with Request %s", path)
		}

	case 304:
		expires, err := RetrieveExpiresHeaderFromResponse(response, 0)
		if err != nil {
			err = errors.Wrapf(err, "Error Encounter with Request %s", path)

			return response, err
		}
		alliance.CachedUntil = expires

		etag, err := RetrieveEtagHeaderFromResponse(response)
		if err != nil {
			err = errors.Wrapf(err, "Error Encounter with Request %s", path)
			return response, err
		}
		alliance.Etag = etag

	}

	response.Data = &alliance

	return response, err
}
