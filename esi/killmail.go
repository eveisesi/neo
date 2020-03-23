package esi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/ddouglas/killboard"
	"github.com/pkg/errors"
)

func (e *Client) GetKillmailsKillmailIDKillmailHash(id, hash string) (Response, error) {
	var response Response

	path := fmt.Sprintf("/v1/killmails/%s/%s/", id, hash)

	url := url.URL{
		Scheme: "https",
		Host:   e.Host,
		Path:   path,
	}

	headers := make(map[string]string)

	request := Request{
		Method:  http.MethodGet,
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

	killmail := killboard.Killmail{}

	err = json.Unmarshal(response.Data.([]byte), &killmail)
	if err != nil {
		err = errors.Wrap(err, "unable to unmarshel response body")
		return response, err
	}

	response.Data = killmail

	return response, err
}
