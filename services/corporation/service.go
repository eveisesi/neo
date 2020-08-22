package corporation

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
	Corporation(ctx context.Context, id uint) (*neo.Corporation, error)
	CorporationsByCorporationIDs(ctx context.Context, ids []uint) ([]*neo.Corporation, error)

	UpdateExpired(ctx context.Context)
}

type service struct {
	redis    *redis.Client
	logger   *logrus.Logger
	newrelic *newrelic.Application
	esi      esi.Service
	tracker  tracker.Service
	neo.CorporationRespository
}

func NewService(redis *redis.Client, logger *logrus.Logger, newrelic *newrelic.Application, esi esi.Service, tracker tracker.Service, corporation neo.CorporationRespository) Service {
	return &service{
		redis,
		logger,
		newrelic,
		esi,
		tracker,
		corporation,
	}
}
