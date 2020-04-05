package neo

import (
	"context"
	"time"

	"github.com/volatiletech/null"
)

type TokenRepository interface {
	Token(ctx context.Context, id uint64) (*Token, error)
	CreateToken(ctx context.Context, token *Token) (*Token, error)
	UpdateToken(ctx context.Context, id uint64, token *Token) (*Token, error)
	DeleteToken(ctx context.Context, id uint64) error
}

type Token struct {
	ID                uint64    `boil:"id" json:"id" toml:"id" yaml:"id"`
	Main              uint64    `boil:"main" json:"main" toml:"main" yaml:"main"`
	AccessToken       string    `boil:"access_token" json:"accessToken" toml:"accessToken" yaml:"accessToken"`
	RefreshToken      string    `boil:"refresh_token" json:"refreshToken" toml:"refreshToken" yaml:"refreshToken"`
	Expiry            time.Time `boil:"expiry" json:"expiry" toml:"expiry" yaml:"expiry"`
	Disabled          bool      `boil:"disabled" json:"disabled" toml:"disabled" yaml:"disabled"`
	DisabledTimestamp null.Time `boil:"disabled_timestamp" json:"disabledTimestamp,omitempty" toml:"disabledTimestamp" yaml:"disabledTimestamp,omitempty"`
	DisabledReason    string    `boil:"disabled_reason" json:"disabledReason" toml:"disabledReason" yaml:"disabledReason"`
	CreatedAt         time.Time `boil:"created_at" json:"createdAt" toml:"createdAt" yaml:"createdAt"`
	UpdatedAt         time.Time `boil:"updated_at" json:"updatedAt" toml:"updatedAt" yaml:"updatedAt"`
}
