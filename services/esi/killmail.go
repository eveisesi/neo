package esi

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/eveisesi/neo"
	"github.com/pkg/errors"
)

func (s *service) GetKillmailsKillmailIDKillmailHash(id uint64, hash string) (*neo.Killmail, *Meta) {

	path := fmt.Sprintf("/v1/killmails/%d/%s/", id, hash)

	request := request{
		method: http.MethodGet,
		path:   path,
	}

	response, m := s.request(request)
	if m.IsError() {
		return nil, m
	}

	killmail := new(neo.Killmail)

	err = json.Unmarshal(response, killmail)
	if err != nil {
		m.Msg = errors.Wrapf(err, "unable to unmarshal response body on request %s", path)
		return nil, m
	}

	killmail.ID = id

	return killmail, m
}
