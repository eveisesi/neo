package market

import (
	"time"

	"github.com/eveisesi/neo"
	"github.com/eveisesi/neo/services/esi"
	"github.com/eveisesi/neo/services/universe"
	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
)

type Service interface {
	FetchHistory(from int)
	FetchTypePrice(id uint64, date time.Time) float64
	neo.MarketRepository
}

type service struct {
	redis    *redis.Client
	esi      esi.Service
	logger   *logrus.Logger
	universe universe.Service
	txn      neo.Starter
	neo.MarketRepository
}

func NewService(redis *redis.Client, esi esi.Service, logger *logrus.Logger, universe universe.Service, txn neo.Starter, market neo.MarketRepository) Service {
	return &service{
		redis,
		esi,
		logger,
		universe,
		txn,
		market,
	}
}
