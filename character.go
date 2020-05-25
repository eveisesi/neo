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
	ID            uint64      `boil:"id" json:"id" toml:"id" yaml:"id"`
	Name          string      `boil:"name" json:"name" toml:"name" yaml:"name"`
	CorporationID uint64      `boil:"corporation_id" json:"corporation_id" toml:"corporation_id" yaml:"corporation_id"`
	AllianceID    null.Uint64 `boil:"alliance_id" json:"alliance_id,omitempty" toml:"alliance_id" yaml:"alliance_id,omitempty"`
	FactionID     null.Uint64 `boil:"faction_id" json:"faction_id,omitempty" toml:"faction_id" yaml:"faction_id,omitempty"`
	Etag          string      `boil:"etag" json:"etag" toml:"etag" yaml:"etag"`
	CachedUntil   time.Time   `boil:"cached_until" json:"cached_until" toml:"cached_until" yaml:"cached_until"`
	CreatedAt     time.Time   `boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`
	UpdatedAt     time.Time   `boil:"updated_at" json:"updated_at" toml:"updated_at" yaml:"updated_at"`
}

func (c Character) IsExpired() bool {
	return c.CachedUntil.Before(time.Now())
}
