package esi

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/eveisesi/neo"
	"github.com/pkg/errors"
)

var TypeNotFound = errors.New("not found")

func (e *Client) GetUniverseTypesTypeID(id uint64) (Response, error) {

	var response Response
	path := fmt.Sprintf("/v3/universe/types/%d/", id)

	url := url.URL{
		Scheme: "https",
		Host:   e.Host,
		Path:   path,
	}

	headers := make(map[string]string)

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

		if response.Code == 404 {
			return response, TypeNotFound
		}

		if response.Code < 400 {
			break
		}

		attempts++
		time.Sleep(time.Second * e.SleepDuration)

	}

	var invType neo.Type
	invType.ID = id

	err = json.Unmarshal(response.Data.([]byte), &invType)
	if err != nil {
		return response, errors.Wrapf(err, "unable to unmarshel response body on request %s", path)
	}

	response.Data = &invType

	return response, err

}
