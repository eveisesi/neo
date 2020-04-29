package esi

import (
	"encoding/json"
	"fmt"
	"net/http"

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
	AttributeID uint64  `json:"attribute_id"`
	Value       float64 `json:"value"`
}

func (s *service) GetUniverseTypesTypeID(id uint64) (*neo.Type, []*neo.TypeAttribute, *Meta) {

	var esitype = new(Type)

	path := fmt.Sprintf("/v3/universe/types/%d/", id)

	request := request{
		method: http.MethodGet,
		path:   path,
	}

	response, m := s.request(request)
	if m.IsError() {
		return nil, nil, m
	}

	err = json.Unmarshal(response, esitype)
	if err != nil {
		m.Msg = errors.Wrapf(err, "unable to unmarshal response body on request %s", path)
		return nil, nil, m
	}

	var attributes = make([]*neo.TypeAttribute, 0)
	for _, v := range esitype.Attributes {
		attributes = append(attributes, &neo.TypeAttribute{
			TypeID:      id,
			AttributeID: v.AttributeID,
			Value:       int64(v.Value),
		})
	}

	return &neo.Type{
		ID:            esitype.ID,
		GroupID:       esitype.GroupID,
		Name:          esitype.Name,
		Description:   esitype.Description,
		Published:     esitype.Published,
		MarketGroupID: esitype.MarketGroupID,
	}, attributes, m

}
