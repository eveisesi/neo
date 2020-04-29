package esi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/eveisesi/neo"
	"github.com/pkg/errors"
	"github.com/volatiletech/null"
)

func (s *service) GetMarketGroups() ([]int, *Meta) {

	response, m := s.request(request{
		method: http.MethodGet,
		path:   "/v1/markets/groups/",
	})
	if m.IsError() {
		return nil, m
	}

	ids := make([]int, 0)
	err := json.Unmarshal(response, &ids)
	if err != nil {
		m.Msg = errors.Wrapf(err, "unable to unmarshal response body on request %s", "/v1/markets/groups/")
		return nil, m
	}

	return ids, m

}

func (s *service) GetMarketGroupsMarketGroupID(id int) (*neo.MarketGroup, *Meta) {

	path := fmt.Sprintf("/v1/markets/groups/%d", id)

	response, m := s.request(request{
		method: http.MethodGet,
		path:   path,
	})
	if m.IsError() {
		return nil, m
	}

	group := new(neo.MarketGroup)
	err := json.Unmarshal(response, group)
	if err != nil {
		m.Msg = errors.Wrapf(err, "unable to unmarshal response body on request %s", path)
		return nil, m
	}

	return group, m

}

func (s *service) GetMarketsRegionIDHistory(regionID uint64, typeID string) ([]*neo.HistoricalRecord, *Meta) {

	path := fmt.Sprintf("/v1/markets/%d/history/", regionID)

	query := url.Values{}
	query.Set("type_id", typeID)

	response, m := s.request(request{
		method: http.MethodGet,
		path:   path,
		query:  query.Encode(),
	})
	if m.IsError() {
		return nil, m
	}

	records := make([]*neo.HistoricalRecord, 0)

	err := json.Unmarshal(response, &records)
	if err != nil {
		m.Msg = errors.Wrapf(err, "unable to unmarshal response body on request %s", path)
		return nil, m
	}

	return records, m
}

func (s *service) HeadMarketsRegionIDTypes(regionID uint64) *Meta {

	_, m := s.request(request{
		method: http.MethodHead,
		path:   fmt.Sprintf("/v1/markets/%d/types/", regionID),
	})
	return m

}

func (s *service) GetMarketsRegionIDTypes(regionID uint64, page null.String) ([]int, *Meta) {

	path := fmt.Sprintf("/v1/markets/%d/types/", regionID)

	query := url.Values{}
	if page.Valid {
		query.Set("page", page.String)
	}

	response, m := s.request(request{
		method: http.MethodGet,
		path:   path,
		query:  query.Encode(),
	})
	if m.IsError() {
		return nil, m
	}

	ids := make([]int, 0)

	err := json.Unmarshal(response, &ids)
	if err != nil {
		m.Msg = errors.Wrapf(err, "unable to unmarshal response body on request %s", path)
		return nil, m
	}

	return ids, m

}
