package neo

import (
	"context"
	"time"

	"github.com/volatiletech/null"
)

type CharacterRespository interface {
	Character(ctx context.Context, id uint64) (*Character, error)
	Characters(ctx context.Context, mods ...Modifier) ([]*Character, error)
	CreateCharacter(ctx context.Context, character *Character) (*Character, error)
	UpdateCharacter(ctx context.Context, id uint64, character *Character) (*Character, error)

	Expired(ctx context.Context) ([]*Character, error)
}

type Character struct {
	ID               uint64      `boil:"id" json:"id"`
	Name             string      `boil:"name" json:"name"`
	CorporationID    uint        `boil:"corporation_id" json:"corporation_id"`
	AllianceID       null.Uint   `boil:"alliance_id" json:"alliance_id,omitempty"`
	FactionID        null.Uint   `boil:"faction_id" json:"faction_id,omitempty"`
	SecurityStatus   float64     `boil:"security_status" json:"security_status"`
	NotModifiedCount uint        `boil:"not_modified_count" json:"not_modified_count"`
	UpdatePriority   uint        `boil:"update_priority" json:"update_priority"`
	Etag             null.String `boil:"etag" json:"etag"`
	CachedUntil      time.Time   `boil:"cached_until" json:"cached_until"`
	CreatedAt        time.Time   `boil:"created_at" json:"created_at"`
	UpdatedAt        time.Time   `boil:"updated_at" json:"updated_at"`
}

func (c Character) IsExpired() bool {
	return c.CachedUntil.Before(time.Now())
}
