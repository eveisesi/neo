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

type (
	SolarSystem struct {
		ID              uint                `json:"system_id"`
		Name            string              `json:"name"`
		ConstellationID uint                `json:"constellation_id"`
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

func (s *service) GetUniverseSystemsSystemID(ctx context.Context, id uint) (*neo.SolarSystem, Meta) {

	var path = fmt.Sprintf("/v4/universe/systems/%d/", id)

	request := request{
		method: http.MethodGet,
		path:   path,
	}

	response, m := s.request(ctx, request)
	if m.IsErr() {
		return nil, m
	}

	var esisystem = new(SolarSystem)

	err = json.Unmarshal(response, esisystem)
	if err != nil {
		m.Msg = errors.Wrap(err, "unable to unmarshal response body")
		return nil, m
	}

	return &neo.SolarSystem{
		ID:              esisystem.ID,
		Name:            esisystem.Name,
		ConstellationID: esisystem.ConstellationID,
		// SunTypeID:       esisystem.SunTypeID,
		// PosX:     esisystem.Position.X,
		// PosY:     esisystem.Position.Y,
		// PosZ:     esisystem.Position.Z,
		Security: esisystem.Security,
	}, m
}
