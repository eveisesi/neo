package neo

import (
	"context"
	"time"

	"github.com/volatiletech/null"
)

type CharacterRespository interface {
	Character(ctx context.Context, id uint64) (*Character, error)
	CreateCharacter(ctx context.Context, character *Character) (*Character, error)
	UpdateCharacter(ctx context.Context, id uint64, character *Character) (*Character, error)
	CharactersByCharacterIDs(ctx context.Context, ids []uint64) ([]*Character, error)

	Expired(ctx context.Context) ([]*Character, error)
}

type Character struct {
	ID               uint64      `boil:"id" json:"id"`
	Name             string      `boil:"name" json:"name"`
	CorporationID    uint64      `boil:"corporation_id" json:"corporation_id"`
	AllianceID       null.Uint64 `boil:"alliance_id" json:"alliance_id,omitempty"`
	FactionID        null.Uint64 `boil:"faction_id" json:"faction_id,omitempty"`
	NotModifiedCount uint        `boil:"not_modified_count" json:"not_modified_count"`
	UpdatePriority   uint        `boil:"update_priority" json:"update_priority"`
	Etag             string      `boil:"etag" json:"etag"`
	CachedUntil      time.Time   `boil:"cached_until" json:"cached_until"`
	CreatedAt        time.Time   `boil:"created_at" json:"created_at"`
	UpdatedAt        time.Time   `boil:"updated_at" json:"updated_at"`
}

func (c Character) IsExpired() bool {
	return c.CachedUntil.Before(time.Now())
}
