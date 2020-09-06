package esi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

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
func (s *service) GetCorporationsCorporationID(ctx context.Context, id uint, etag null.String) (*neo.Corporation, Meta) {

	path := fmt.Sprintf("/v4/corporations/%d/", id)
	headers := make(map[string]string)

	if etag.Valid {
		headers["If-None-Match"] = etag.String
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

	corporation := new(neo.Corporation)

	switch m.Code {
	case 200:
		err := json.Unmarshal(response, corporation)
		if err != nil {
			m.Msg = errors.Wrapf(err, "unable to unmarshal response body on request %s", path)
			return nil, m
		}

		corporation.ID = id

	}
	corporation.CachedUntil = s.retrieveExpiresHeader(m.Headers, 0).Unix()
	if s.retrieveEtagHeader(m.Headers) != "" {
		corporation.Etag = s.retrieveEtagHeader(m.Headers)
	}
	return corporation, m

}
