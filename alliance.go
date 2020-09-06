package neo

import (
	"context"
)

type AllianceRespository interface {
	Alliance(ctx context.Context, id uint) (*Alliance, error)
	Alliances(ctx context.Context, mods ...Modifier) ([]*Alliance, error)
	CreateAlliance(ctx context.Context, alliance *Alliance) error
	UpdateAlliance(ctx context.Context, id uint, alliance *Alliance) error

	Expired(ctx context.Context) ([]*Alliance, error)
}

// Alliance is an object representing the database table.
type Alliance struct {
	ID               uint   `bson:"id" json:"id"`
	Name             string `bson:"name" json:"name"`
	Ticker           string `bson:"ticker" json:"ticker"`
	IsClosed         bool   `bson:"isClosed" json:"isClosed"`
	NotModifiedCount uint   `bson:"notModifiedCount" json:"notModifiedCount"`
	UpdatePriority   uint   `bson:"updatePriority" json:"updatePriority"`
	Etag             string `bson:"etag" json:"etag"`
	CachedUntil      int64  `bson:"cachedUntil" json:"cachedUntil"`
	CreatedAt        int64  `bson:"createdAt" json:"createdAt"`
	UpdatedAt        int64  `bson:"updatedAt" json:"updatedAt"`
}
