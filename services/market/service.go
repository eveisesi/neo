package market

import (
	"time"

	"github.com/eveisesi/neo"
	"github.com/eveisesi/neo/services/esi"
	"github.com/eveisesi/neo/services/tracker"
	"github.com/eveisesi/neo/services/universe"
	"github.com/go-redis/redis/v7"
	"github.com/sirupsen/logrus"
)

type Service interface {
	FetchHistory()
	FetchTypePrice(id uint64, date time.Time) float64
	FetchPrices()
	neo.MarketRepository
}

type service struct {
	redis    *redis.Client
	esi      esi.Service
	logger   *logrus.Logger
	universe universe.Service
	txn      neo.Starter
	neo.MarketRepository
	tracker tracker.Service
}

func NewService(redis *redis.Client, esi esi.Service, logger *logrus.Logger, universe universe.Service, txn neo.Starter, market neo.MarketRepository, tracker tracker.Service) Service {
	return &service{
		redis,
		esi,
		logger,
		universe,
		txn,
		market,
		tracker,
	}
}
