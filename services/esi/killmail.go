package esi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/eveisesi/neo"
	"github.com/pkg/errors"
)

func (s *service) GetKillmailsKillmailIDKillmailHash(id, hash string) (*neo.Killmail, *Meta) {

	path := fmt.Sprintf("/v1/killmails/%s/%s/", id, hash)

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

	u, _ := strconv.ParseUint(id, 10, 64)

	killmail.ID = uint64(u)

	return killmail, m
}
