package neo

import (
	"context"
)

type CorporationRespository interface {
	Corporation(ctx context.Context, id uint) (*Corporation, error)
	Corporations(ctx context.Context, mods ...Modifier) ([]*Corporation, error)
	CreateCorporation(ctx context.Context, corporation *Corporation) error
	UpdateCorporation(ctx context.Context, id uint, corporation *Corporation) error

	Expired(ctx context.Context) ([]*Corporation, error)
	MemberCountByAllianceID(ctx context.Context, id uint) (int, error)
}

type Corporation struct {
	ID               uint   `bson:"id" json:"id"`
	Name             string `bson:"name" json:"name"`
	Ticker           string `bson:"ticker" json:"ticker"`
	MemberCount      uint   `bson:"memberCount" json:"memberCount"`
	AllianceID       *uint  `bson:"allianceID,omitempty" json:"allianceID,omitempty"`
	NotModifiedCount uint   `bson:"notModifiedCount" json:"notModifiedCount"`
	UpdatePriority   uint   `bson:"updatePriority" json:"updatePriority"`
	Etag             string `bson:"etag" json:"etag"`
	CachedUntil      int64  `bson:"cachedUntil" json:"cachedUntil"`
	UpdateError      int64  `bson:"updateError" json:"updateError"`
	CreatedAt        int64  `bson:"createdAt" json:"createdAt"`
	UpdatedAt        int64  `bson:"updatedAt" json:"updatedAt"`
}
