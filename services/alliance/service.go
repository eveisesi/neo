package alliance

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
	AlliancesByAllianceIDs(ctx context.Context, ids []uint) ([]*neo.Alliance, error)
	neo.AllianceRespository
}

type service struct {
	redis    *redis.Client
	logger   *logrus.Logger
	newrelic *newrelic.Application
	esi      esi.Service
	tracker  tracker.Service
	neo.AllianceRespository
}

func NewService(redis *redis.Client, logger *logrus.Logger, newrelic *newrelic.Application, esi esi.Service, tracker tracker.Service, alliance neo.AllianceRespository) Service {
	return &service{
		redis,
		logger,
		newrelic,
		esi,
		tracker,
		alliance,
	}
}
