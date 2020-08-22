package neo

import (
	"context"
	"time"

	"github.com/volatiletech/null"
)

type CorporationRespository interface {
	Corporation(ctx context.Context, id uint) (*Corporation, error)
	CreateCorporation(ctx context.Context, corporation *Corporation) (*Corporation, error)
	UpdateCorporation(ctx context.Context, id uint, corporation *Corporation) (*Corporation, error)
	CorporationsByCorporationIDs(ctx context.Context, ids []uint) ([]*Corporation, error)

	Expired(ctx context.Context) ([]*Corporation, error)
}

type Corporation struct {
	ID               uint        `boil:"id" json:"id"`
	Name             string      `boil:"name" json:"name"`
	Ticker           string      `boil:"ticker" json:"ticker"`
	MemberCount      uint        `boil:"member_count" json:"member_count"`
	AllianceID       null.Uint `boil:"alliance_id" json:"alliance_id,omitempty"`
	NotModifiedCount uint        `boil:"not_modified_count" json:"not_modified_count"`
	UpdatePriority   uint        `boil:"update_priority" json:"update_priority"`
	Etag             null.String `boil:"etag" json:"etag"`
	CachedUntil      time.Time   `boil:"cached_until" json:"cached_until"`
	CreatedAt        time.Time   `boil:"created_at" json:"created_at"`
	UpdatedAt        time.Time   `boil:"updated_at" json:"updated_at"`
}

func (c Corporation) IsExpired() bool {
	return c.CachedUntil.Before(time.Now())
}
