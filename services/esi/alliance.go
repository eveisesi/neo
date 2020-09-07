package esi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/eveisesi/neo"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
)

// Alliance is an object representing the database table.
type Alliance struct {
	ID     uint   `json:"id"`
	Name   string `json:"name"`
	Ticker string `json:"ticker"`
}

func (r Alliance) validate() bool {
	if r.Name == "" || r.Ticker == "" {
		return false
	}
	return true
}

// GetAlliancesAllianceID makes a HTTP GET Request to the /alliances/{alliance_id} endpoint
// for information about the provided alliance
//
// Documentation: https://esi.evetech.net/ui/#/Alliance/get_alliances_alliance_id
// Version: v3
// Cache: 3600 sec (1 Hour)
func (s *service) GetAlliancesAllianceID(ctx context.Context, id uint, etag string) (*neo.Alliance, Meta) {

	path := fmt.Sprintf("/v3/alliances/%d/", id)
	headers := make(map[string]string)

	if etag != "" {
		headers["If-None-Match"] = etag
	}

	request := request{
		method:  http.MethodGet,
		path:    path,
		headers: headers,
	}

	response, m := s.request(ctx, request)
	if m.IsErr() {
		return nil, m
	}

	esiAlliance := new(Alliance)

	switch m.Code {
	case 200:
		err = json.Unmarshal(response, esiAlliance)
		if err != nil {
			m.Msg = errors.Wrapf(err, "unable to unmarshal response body on request %s", path)
			return nil, m
		}

		esiAlliance.ID = id

		if !esiAlliance.validate() {
			m.Msg = errors.New("invalid data received from ESI.")
			m.Code = http.StatusUnprocessableEntity
			return nil, m
		}

	}

	alliance := new(neo.Alliance)
	err = copier.Copy(alliance, esiAlliance)
	if err != nil {
		m.Msg = err
		return nil, m
	}

	alliance.CachedUntil = s.retrieveExpiresHeader(m.Headers, 0).Unix()
	if s.retrieveEtagHeader(m.Headers) != "" {
		alliance.Etag = s.retrieveEtagHeader(m.Headers)
	}

	return alliance, m
}
