package neo

import (
	"context"
)

type CharacterRespository interface {
	Character(ctx context.Context, id uint64) (*Character, error)
	Characters(ctx context.Context, operators ...*Operator) ([]*Character, error)
	CreateCharacter(ctx context.Context, character *Character) error
	UpdateCharacter(ctx context.Context, id uint64, character *Character) error
	DeleteCharacter(ctx context.Context, id uint64) error

	Expired(ctx context.Context) ([]*Character, error)
}

type Character struct {
	ID               uint64  `bson:"id" json:"id"`
	Name             string  `bson:"name" json:"name"`
	CorporationID    uint    `bson:"corporationID" json:"corporationID"`
	AllianceID       *uint   `bson:"allianceID,omitempty" json:"allianceID,omitempty"`
	FactionID        *uint   `bson:"factionID,omitempty" json:"factionID,omitempty"`
	SecurityStatus   float64 `bson:"securityStatus" json:"securityStatus"`
	NotModifiedCount uint    `bson:"notModifiedCount" json:"notModifiedCount"`
	UpdatePriority   uint    `bson:"updatePriority" json:"updatePriority"`
	Etag             string  `bson:"etag" json:"etag"`
	CachedUntil      int64   `bson:"cachedUntil" json:"cachedUntil"`
	UpdateError      int64   `bson:"updateError" json:"updateError"`
	CreatedAt        int64   `bson:"createdAt" json:"createdAt"`
	UpdatedAt        int64   `bson:"updatedAt" json:"updatedAt"`
}
