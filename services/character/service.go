package character

import (
	"context"

	"github.com/eveisesi/neo"
	"github.com/eveisesi/neo/esi"
	"github.com/go-redis/redis"
)

type Service interface {
	Character(ctx context.Context, id uint64) (*neo.Character, error)
	CharactersByCharacterIDs(ctx context.Context, ids []uint64) ([]*neo.Character, error)
}

type service struct {
	redis *redis.Client
	esi   *esi.Client
	neo.CharacterRespository
}

func NewService(redis *redis.Client, esi *esi.Client, character neo.CharacterRespository) Service {
	return &service{
		redis,
		esi,
		character,
	}
}
