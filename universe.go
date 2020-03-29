package killboard

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
	GroupID       uint64       `json:"group_id"`
	Name          string       `json:"name"`
	Description   string       `json:"description"`
	Volume        float64      `json:"volume"`
	RaceID        null.Uint64  `json:"race_id"`
	BasePrice     null.Float64 `json:"base_price"`
	Published     bool         `json:"published"`
	MarketGroupID null.Uint64  `json:"market_group_id"`
	CreatedAt     null.Time    `json:"created_at"`
	UpdatedAt     null.Time    `json:"updated_at"`
}

type SolarSystem struct {
	ID              uint64      `json:"id"`
	Name            null.String `json:"name"`
	RegionID        uint64      `json:"region_id"`
	ConstellationID uint64      `json:"constellation_id"`
	FactionID       null.Int64  `json:"faction_id"`
	SunTypeID       null.Int64  `json:"sun_type_id"`
	PosX            float64     `json:"pos_x"`
	PosY            float64     `json:"pos_y"`
	PosZ            float64     `json:"pos_z"`
	Security        float64     `json:"security"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}
