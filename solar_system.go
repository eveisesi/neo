package killboard

import (
	"time"

	"github.com/volatiletech/null"
)

type SolarSystem struct {
	ID              int64       `json:"id"`
	Name            null.String `json:"name"`
	RegionID        int64       `json:"region_id"`
	ConstellationID int64       `json:"constellation_id"`
	FactionID       null.Int64  `json:"faction_id"`
	SunTypeID       null.Int64  `json:"sun_type_id"`
	PosX            float64     `json:"pos_x"`
	PosY            float64     `json:"pos_y"`
	PosZ            float64     `json:"pos_z"`
	Security        float64     `json:"security"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}
