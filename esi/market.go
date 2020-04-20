package esi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/volatiletech/null"
)

func (e *Client) HeadMarketsRegionIDOrders(regionID uint64) (Response, error) {

	var response Response
	path := fmt.Sprintf("/v1/markets/%d/orders/", regionID)

	query := url.Values{}
	query.Set("order_type", "sell")
	query.Set("type_id", "33602")

	url := url.URL{
		Scheme:   "https",
		Host:     "esi.evetech.net",
		Path:     path,
		RawQuery: query.Encode(),
	}

	request := Request{
		Method: http.MethodHead,
		Path:   url,
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

		if response.Code < 400 {
			break
		}

		attempts++
		time.Sleep(time.Second * e.SleepDuration)

	}

	return response, err

}

type Order struct {
	OrderID    uint64  `json:"order_id"`
	LocationID uint    `json:"location_id"`
	SystemID   uint    `json:"system_id"`
	TypeID     uint    `json:"type_id"`
	Price      float64 `json:"price"`
}

func (e *Client) GetMarketsRegionIDOrders(regionID uint64, page null.Uint) (Response, error) {

	var response Response
	path := fmt.Sprintf("/v1/markets/%d/orders/", regionID)

	query := url.Values{}
	query.Set("order_type", "sell")
	query.Set("type_id", "33602")
	if page.Valid {
		strPage := strconv.FormatUint(uint64(page.Uint), 10)
		query.Set("page", strPage)
	}

	url := url.URL{
		Scheme:   "https",
		Host:     "esi.evetech.net",
		Path:     path,
		RawQuery: query.Encode(),
	}

	request := Request{
		Method: http.MethodGet,
		Path:   url,
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

		if response.Code < 400 {
			break
		}

		attempts++
		time.Sleep(time.Second * e.SleepDuration)

	}

	var orders = make([]*Order, 0)

	err := json.Unmarshal(response.Data.([]byte), &orders)
	if err != nil {
		return response, errors.Wrap(err, "failed to unmarshal response body")
	}

	response.Data = orders

	return response, err

}
