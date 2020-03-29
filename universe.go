package neo

import (
	"context"
	"time"

	"github.com/volatiletech/null"
)

type UniverseRepository interface {
	Type(ctx context.Context, id uint64) (*Type, error)
	TypesByTypeIDs(ctx context.Context, ids []uint64) ([]*Type, error)
	SolarSystem(ctx context.Context, id uint64) (*SolarSystem, error)
	SolarSystemsBySolarSystemIDs(ctx context.Context, ids []uint64) ([]*SolarSystem, error)
}

// Type is an object representing the database table.
type Type struct {
	ID            uint64       `json:"id"`
	GroupID       uint64       `json:"groupID"`
	Name          string       `json:"name"`
	Description   string       `json:"description"`
	Volume        float64      `json:"volume"`
	RaceID        null.Uint64  `json:"raceID"`
	BasePrice     null.Float64 `json:"base_price"`
	Published     bool         `json:"published"`
	MarketGroupID null.Uint64  `json:"market_groupID"`
	CreatedAt     null.Time    `json:"created_at"`
	UpdatedAt     null.Time    `json:"updated_at"`
}

type SolarSystem struct {
	ID              uint64     `json:"id"`
	Name            string     `json:"name"`
	RegionID        uint64     `json:"regionID"`
	ConstellationID uint64     `json:"constellationID"`
	FactionID       null.Int64 `json:"factionID"`
	SunTypeID       null.Int64 `json:"sun_typeID"`
	PosX            float64    `json:"pos_x"`
	PosY            float64    `json:"pos_y"`
	PosZ            float64    `json:"pos_z"`
	Security        float64    `json:"security"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}
