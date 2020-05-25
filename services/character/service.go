package character

import (
	"context"

	"github.com/eveisesi/neo"
	"github.com/eveisesi/neo/services/esi"
	"github.com/eveisesi/neo/services/tracker"
	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
)

type Service interface {
	Character(ctx context.Context, id uint64) (*neo.Character, error)
	CharactersByCharacterIDs(ctx context.Context, ids []uint64) ([]*neo.Character, error)

	UpdateExpired(ctx context.Context)
}

type service struct {
	redis   *redis.Client
	logger  *logrus.Logger
	esi     esi.Service
	tracker tracker.Service
	neo.CharacterRespository
}

func NewService(redis *redis.Client, logger *logrus.Logger, esi esi.Service, tracker tracker.Service, character neo.CharacterRespository) Service {
	return &service{
		redis,
		logger,
		esi,
		tracker,
		character,
	}
}
