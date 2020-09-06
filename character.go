package neo

import (
	"context"
)

type CharacterRespository interface {
	Character(ctx context.Context, id uint64) (*Character, error)
	Characters(ctx context.Context, mods ...Modifier) ([]*Character, error)
	CreateCharacter(ctx context.Context, character *Character) error
	UpdateCharacter(ctx context.Context, id uint64, character *Character) error
	DeleteCharacter(ctx context.Context, id uint64) error

	Expired(ctx context.Context) ([]*Character, error)
}

type Character struct {
	ID               uint64  `db:"id" bson:"id" json:"id"`
	Name             string  `db:"name" bson:"name" json:"name"`
	CorporationID    uint    `db:"corporationID" bson:"corporationID" json:"corporationID"`
	AllianceID       *uint   `db:"allianceID" bson:"allianceID,omitempty" json:"allianceID,omitempty"`
	FactionID        *uint   `db:"factionID" bson:"factionID,omitempty" json:"factionID,omitempty"`
	SecurityStatus   float64 `db:"securityStatus" bson:"securityStatus" json:"securityStatus"`
	NotModifiedCount uint    `db:"notModifiedCount" bson:"notModifiedCount" json:"notModifiedCount"`
	UpdatePriority   uint    `db:"updatePriority" bson:"updatePriority" json:"updatePriority"`
	Etag             string  `db:"etag" bson:"etag" json:"etag"`
	CachedUntil      int64   `db:"cachedUntil" bson:"cachedUntil" json:"cachedUntil"`
	CreatedAt        int64   `db:"createdAt" bson:"createdAt" json:"createdAt"`
	UpdatedAt        int64   `db:"updatedAt" bson:"updatedAt" json:"updatedAt"`
}
