package neo

import (
	"context"
	"time"
)

type AllianceRespository interface {
	Alliance(ctx context.Context, id uint64) (*Alliance, error)
	CreateAlliance(ctx context.Context, alliance *Alliance) (*Alliance, error)
	UpdateAlliance(ctx context.Context, id uint64, alliance *Alliance) (*Alliance, error)
	AlliancesByAllianceIDs(ctx context.Context, ids []uint64) ([]*Alliance, error)
}

// Alliance is an object representing the database table.
type Alliance struct {
	ID          uint64    `json:"id"`
	Name        string    `json:"name"`
	Ticker      string    `json:"ticker"`
	MemberCount uint64    `json:"member_count"`
	IsClosed    int8      `json:"is_closed"`
	Etag        string    `json:"etag"`
	CachedUntil time.Time `json:"cached_until"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (a Alliance) IsExpired() bool {
	return a.CachedUntil.Before(time.Now())
}
