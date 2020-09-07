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

type Corporation struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Ticker      string `json:"ticker"`
	MemberCount uint   `json:"member_count"`
	AllianceID  *uint  `json:"alliance_id,omitempty"`
}

func (r Corporation) validate() bool {
	if r.Name == "" || r.Ticker == "" {
		return false
	}

	return true
}

// GetCorporationsCorporationID makes a HTTP GET Request to the /corporations/{corporation_id} endpoint
// for information about the provided corporation
//
// Documentation: https://esi.evetech.net/ui/#/Corporation/get_corporations_corporation_id
// Version: v4
// Cache: 3600 sec (1 Hour)
func (s *service) GetCorporationsCorporationID(ctx context.Context, id uint, etag string) (*neo.Corporation, Meta) {

	path := fmt.Sprintf("/v4/corporations/%d/", id)
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

	esiCorporation := new(Corporation)

	switch m.Code {
	case 200:
		err := json.Unmarshal(response, esiCorporation)
		if err != nil {
			m.Msg = errors.Wrapf(err, "unable to unmarshal response body on request %s", path)
			return nil, m
		}

		esiCorporation.ID = id
		if !esiCorporation.validate() {
			m.Msg = errors.New("invalid data received from ESI.")
			m.Code = http.StatusUnprocessableEntity
			return nil, m
		}

	}

	corporation := new(neo.Corporation)
	err = copier.Copy(corporation, esiCorporation)
	if err != nil {
		m.Msg = err
		return nil, m
	}

	corporation.CachedUntil = s.retrieveExpiresHeader(m.Headers, 0).Unix()
	if s.retrieveEtagHeader(m.Headers) != "" {
		corporation.Etag = s.retrieveEtagHeader(m.Headers)
	}
	return corporation, m

}
