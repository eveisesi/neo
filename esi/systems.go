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

type (
	SolarSystem struct {
		ID              uint64              `json:"system_id"`
		Name            string              `json:"name"`
		ConstellationID uint64              `json:"constellation_id"`
		SunTypeID       null.Int64          `json:"star_id"`
		Position        SolarSystemPosition `json:"position"`
		Security        float64             `json:"security_status"`
	}

	SolarSystemPosition struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
		Z float64 `json:"z"`
	}
)

func (e *Client) GetUniverseSystemsSystemID(id uint64) (Response, error) {

	var response Response
	path := fmt.Sprintf("/v4/universe/systems/%d/", id)

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
			return response, neo.ErrEsiMaxAttempts
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

	var esisystem = new(SolarSystem)

	err = json.Unmarshal(response.Data.([]byte), esisystem)
	if err != nil {
		err = errors.Wrap(err, "unable to unmarshel response body")
		return response, err
	}

	response.Data = &neo.SolarSystem{
		ID:              esisystem.ID,
		Name:            esisystem.Name,
		ConstellationID: esisystem.ConstellationID,
		SunTypeID:       esisystem.SunTypeID,
		PosX:            esisystem.Position.X,
		PosY:            esisystem.Position.Y,
		PosZ:            esisystem.Position.Z,
		Security:        esisystem.Security,
	}

	return response, err
}
