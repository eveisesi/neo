package alliance

import (
	"context"

	"github.com/eveisesi/neo"
	"github.com/eveisesi/neo/services/esi"
	"github.com/eveisesi/neo/services/tracker"
	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
)

type Service interface {
	Alliance(ctx context.Context, id uint64) (*neo.Alliance, error)
	AlliancesByAllianceIDs(ctx context.Context, ids []uint64) ([]*neo.Alliance, error)

	UpdateExpired(ctx context.Context)
}

type service struct {
	redis   *redis.Client
	logger  *logrus.Logger
	esi     esi.Service
	tracker tracker.Service
	neo.AllianceRespository
}

func NewService(redis *redis.Client, logger *logrus.Logger, esi esi.Service, tracker tracker.Service, alliance neo.AllianceRespository) Service {
	return &service{
		redis,
		logger,
		esi,
		tracker,
		alliance,
	}
}
