package esi

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/eveisesi/neo"
	"github.com/pkg/errors"
	"github.com/volatiletech/null"
)

// GetAlliancesAllianceID makes a HTTP GET Request to the /alliances/{alliance_id} endpoint
// for information about the provided alliance
//
// Documentation: https://esi.evetech.net/ui/#/Alliance/get_alliances_alliance_id
// Version: v3
// Cache: 3600 sec (1 Hour)
func (s *service) GetAlliancesAllianceID(id uint64, etag null.String) (*neo.Alliance, *Meta) {

	path := fmt.Sprintf("/v3/alliances/%d/", id)
	headers := make(map[string]string)

	if etag.Valid {
		headers["If-None-Match"] = etag.String
	}

	request := request{
		method:  http.MethodGet,
		path:    path,
		headers: headers,
	}

	response, m := s.request(request)
	if m.IsError() {
		return nil, m
	}

	alliance := new(neo.Alliance)

	switch m.Code {
	case 200:
		err = json.Unmarshal(response, alliance)
		if err != nil {
			m.Msg = errors.Wrapf(err, "unable to unmarshal response body on request %s", path)
			return nil, m
		}

		alliance.ID = id

	}

	alliance.CachedUntil = s.retrieveExpiresHeader(m.Headers, 0)
	alliance.Etag = s.retrieveEtagHeader(m.Headers)

	return alliance, m
}
