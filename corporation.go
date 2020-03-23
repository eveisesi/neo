package killboard

import (
	"time"

	"github.com/volatiletech/null"
)

type Corporation struct {
	ID            uint64      `db:"id" json:"id"`
	Name          string      `db:"name" json:"name"`
	Ticker        string      `db:"ticker" json:"ticker"`
	MemberCount   uint64      `db:"member_count" json:"member_count"`
	CeoID         uint64      `db:"ceo_id" json:"ceo_id"`
	AllianceID    null.Uint64 `db:"alliance_id" json:"alliance_id,omitempty"`
	DateFounded   null.Time   `db:"date_founded" json:"date_founded,omitempty"`
	CreatorID     uint64      `db:"creator_id" json:"creator_id"`
	HomeStationID null.Uint64 `db:"home_station_id" json:"home_station_id,omitempty"`
	TaxRate       float64     `db:"tax_rate" json:"tax_rate"`
	WarEligible   bool        `db:"war_eligible" json:"war_eligible"`
	Ignored       bool        `db:"ignored" json:"ignored"`
	Closed        bool        `db:"closed" json:"closed"`
	Etag          string      `db:"etag" json:"etag"`
	Expires       time.Time   `db:"expires" json:"expires"`
	CreatedAt     time.Time   `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time   `db:"updated_at" json:"updated_at"`
}

func (c Corporation) IsExpired() bool {
	return c.Expires.Before(time.Now())
}
