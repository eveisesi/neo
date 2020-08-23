package character

import (
	"context"

	"github.com/eveisesi/neo"
	"github.com/eveisesi/neo/services/esi"
	"github.com/eveisesi/neo/services/tracker"
	"github.com/go-redis/redis/v7"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/sirupsen/logrus"
)

type Service interface {
	UpdateExpired(ctx context.Context)
	CharactersByCharacterIDs(ctx context.Context, ids []uint64) ([]*neo.Character, error)
	neo.CharacterRespository
}

type service struct {
	redis    *redis.Client
	logger   *logrus.Logger
	newrelic *newrelic.Application
	esi      esi.Service
	tracker  tracker.Service
	neo.CharacterRespository
}

func NewService(redis *redis.Client, logger *logrus.Logger, newrelic *newrelic.Application, esi esi.Service, tracker tracker.Service, character neo.CharacterRespository) Service {
	return &service{
		redis,
		logger,
		newrelic,
		esi,
		tracker,
		character,
	}
}
