package esi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/eveisesi/neo"
	"github.com/pkg/errors"
	"github.com/volatiletech/null"
)

// Type is an object representing the database table.
type Type struct {
	ID            uint64           `json:"type_id"`
	GroupID       uint64           `json:"group_id"`
	Name          string           `json:"name"`
	Description   string           `json:"description"`
	Published     bool             `json:"published"`
	MarketGroupID null.Uint64      `json:"marketGroupID"`
	Attributes    []*TypeAttribute `json:"dogma_attributes"`
}

type TypeAttribute struct {
	AttributeID uint64 `json:"attribute_id"`
	Value       int64  `json:"value"`
}

func (e *Client) GetUniverseTypesTypeID(id uint64) (Response, error) {

	var esitype = new(Type)
	var response Response
	path := fmt.Sprintf("/v3/universe/types/%d/", id)

	url := url.URL{
		Scheme: "https",
		Host:   e.Host,
		Path:   path,
	}

	request := Request{
		Method:  http.MethodGet,
		Path:    url,
		Headers: make(map[string]string),
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
			return response, neo.ErrEsiTypeNotFound
		}

		if response.Code < 400 {
			break
		}

		attempts++
		time.Sleep(time.Second * e.SleepDuration)

	}

	err = json.Unmarshal(response.Data.([]byte), esitype)
	if err != nil {
		return response, errors.Wrapf(err, "unable to unmarshel response body on request %s", path)
	}

	var attributes = make([]*neo.TypeAttribute, 0)
	for _, v := range esitype.Attributes {
		attributes = append(attributes, &neo.TypeAttribute{
			TypeID:      id,
			AttributeID: v.AttributeID,
			Value:       v.Value,
		})
	}

	response.Data = map[string]interface{}{
		"type": &neo.Type{
			ID:            esitype.ID,
			GroupID:       esitype.GroupID,
			Name:          esitype.Name,
			Description:   esitype.Description,
			Published:     esitype.Published,
			MarketGroupID: esitype.MarketGroupID,
		},
		"attributes": attributes,
	}
	return response, err

}
