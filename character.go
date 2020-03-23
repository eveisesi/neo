package killboard

import (
	"time"

	"github.com/volatiletech/null"
)

type Character struct {
	ID             uint64      `db:"id" json:"id"`
	Name           string      `db:"name" json:"name"`
	Birthday       null.Time   `db:"birthday" json:"birthday,omitempty"`
	Gender         string      `db:"gender" json:"gender"`
	SecurityStatus float64     `db:"security_status" json:"security_status"`
	AllianceID     null.Uint64 `db:"alliance_id" json:"alliance_id,omitempty"`
	CorporationID  uint64      `db:"corporation_id" json:"corporation_id"`
	FactionID      null.Uint64 `db:"faction_id" json:"faction_id,omitempty"`
	AncestryID     uint64      `db:"ancestry_id" json:"ancestry_id"`
	BloodlineID    uint64      `db:"bloodline_id" json:"bloodline_id"`
	RaceID         uint64      `db:"race_id" json:"race_id"`
	Ignored        bool        `db:"ignored" json:"ignored"`
	Etag           string      `db:"etag" json:"etag"`
	Expires        time.Time   `db:"expires" json:"expires"`
	CreatedAt      time.Time   `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time   `db:"updated_at" json:"updated_at"`
}

func (c Character) IsExpired() bool {
	return c.Expires.Before(time.Now())
}
