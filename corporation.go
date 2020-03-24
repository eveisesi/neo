package killboard

import (
	"time"
)

type Corporation struct {
	ID          uint64    `boil:"id" json:"id" toml:"id" yaml:"id"`
	Name        string    `boil:"name" json:"name" toml:"name" yaml:"name"`
	Ticker      string    `boil:"ticker" json:"ticker" toml:"ticker" yaml:"ticker"`
	AllianceID  uint64    `boil:"alliance_id" json:"alliance_id" toml:"alliance_id" yaml:"alliance_id"`
	Etag        string    `boil:"etag" json:"etag" toml:"etag" yaml:"etag"`
	CachedUntil time.Time `boil:"cached_until" json:"cached_until" toml:"cached_until" yaml:"cached_until"`
	CreatedAt   time.Time `boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`
	UpdatedAt   time.Time `boil:"updated_at" json:"updated_at" toml:"updated_at" yaml:"updated_at"`
}

func (c Corporation) IsExpired() bool {
	return c.CachedUntil.Before(time.Now())
}
