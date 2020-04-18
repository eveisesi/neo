package esi

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/eveisesi/neo"
	"github.com/pkg/errors"
	"github.com/volatiletech/null"
)

// GetCorporationsCorporationID makes a HTTP GET Request to the /corporations/{corporation_id} endpoint
// for information about the provided corporation
//
// Documentation: https://esi.evetech.net/ui/#/Corporation/get_corporations_corporation_id
// Version: v4
// Cache: 3600 sec (1 Hour)
func (e *Client) GetCorporationsCorporationID(id uint64, etag null.String) (Response, error) {
	var response Response
	path := fmt.Sprintf("/v4/corporations/%d/", id)

	url := url.URL{
		Scheme: "https",
		Host:   e.Host,
		Path:   path,
	}

	headers := make(map[string]string)

	if etag.Valid {
		headers["If-None-Match"] = etag.String
	}

	request := Request{
		Method:  "GET",
		Path:    url,
		Headers: headers,
		Body:    []byte(""),
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

	var corporation neo.Corporation

	switch response.Code {
	case 200:

		err := json.Unmarshal(response.Data.([]byte), &corporation)
		if err != nil {
			return response, errors.Wrap(err, "unable to unmarshel response body")
		}

		corporation.ID = id

		corporation.CachedUntil, err = RetrieveExpiresHeaderFromResponse(response, 0)
		if err != nil {
			return response, errors.Wrap(err, "Error Encountered attempting to parse expires header")
		}

		corporation.Etag, err = RetrieveEtagHeaderFromResponse(response)
		if err != nil {
			return response, errors.Wrap(err, "Error Encountered attempting to retrieve etag header")
		}

	case 304:
		corporation.CachedUntil, err = RetrieveExpiresHeaderFromResponse(response, 0)
		if err != nil {
			return response, errors.Wrap(err, "Error Encountered attempting to parse expires header")
		}

		corporation.Etag, err = RetrieveEtagHeaderFromResponse(response)
		if err != nil {
			return response, errors.Wrap(err, "Error Encountered attempting to retrieve etag header")
		}

	}

	response.Data = &corporation

	return response, err

}
