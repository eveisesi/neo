package neo

import (
	"context"
	"time"

	"github.com/volatiletech/null"
)

type AllianceRespository interface {
	Alliance(ctx context.Context, id uint64) (*Alliance, error)
	CreateAlliance(ctx context.Context, alliance *Alliance) (*Alliance, error)
	UpdateAlliance(ctx context.Context, id uint64, alliance *Alliance) (*Alliance, error)
	AlliancesByAllianceIDs(ctx context.Context, ids []uint64) ([]*Alliance, error)

	Expired(ctx context.Context) ([]*Alliance, error)
}

// Alliance is an object representing the database table.
type Alliance struct {
	ID               uint64      `boil:"id" json:"id"`
	Name             string      `boil:"name" json:"name"`
	Ticker           string      `boil:"ticker" json:"ticker"`
	MemberCount      uint64      `boil:"member_count" json:"member_count"`
	IsClosed         bool        `boil:"is_closed" json:"is_closed"`
	NotModifiedCount uint        `boil:"not_modified_count" json:"not_modified_count"`
	UpdatePriority   uint        `boil:"update_priority" json:"update_priority"`
	Etag             null.String `boil:"etag" json:"etag"`
	CachedUntil      time.Time   `boil:"cached_until" json:"cached_until"`
	CreatedAt        time.Time   `boil:"created_at" json:"created_at"`
	UpdatedAt        time.Time   `boil:"updated_at" json:"updated_at"`
}

func (a Alliance) IsExpired() bool {
	return a.CachedUntil.Before(time.Now())
}
