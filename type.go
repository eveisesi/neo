package killboard

import (
	"github.com/volatiletech/null"
)

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
