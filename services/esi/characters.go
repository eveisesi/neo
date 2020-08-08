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

// GetCharactersCharacterID makes a HTTP GET Request to the /characters/{character_id} endpoint
// for information about the provided character
//
// Documentation: https://esi.evetech.net/ui/#/Character/get_characters_character_id
// Version: v4
// Cache: 86400 sec (24 Hour)
func (s *service) GetCharactersCharacterID(ctx context.Context, id uint64, etag null.String) (*neo.Character, *Meta) {

	path := fmt.Sprintf("/v4/characters/%d/", id)
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
	if m.IsError() {
		return nil, m
	}

	character := new(neo.Character)

	switch m.Code {
	case 200:
		err := json.Unmarshal(response, character)
		if err != nil {
			m.Msg = errors.Wrapf(err, "unable to unmarshal response body on request %s", path)
			return nil, m
		}

		character.ID = id

	}
	character.CachedUntil = s.retrieveExpiresHeader(m.Headers, 0)
	character.Etag.SetValid(s.retrieveEtagHeader(m.Headers))

	return character, m
}
