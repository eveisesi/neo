package market

import (
	"time"

	"github.com/eveisesi/neo"
	"github.com/eveisesi/neo/esi"
	"github.com/eveisesi/neo/services/universe"
	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
)

type Service interface {
	FetchOrders()
	CalculateRawMaterialCost(id uint64, minDate, maxDate time.Time) float64
	neo.MarketRepository
}

type service struct {
	redis    *redis.Client
	esi      *esi.Client
	logger   *logrus.Logger
	universe universe.Service
	txn      neo.Starter
	neo.MarketRepository
}

func NewService(redis *redis.Client, esi *esi.Client, logger *logrus.Logger, universe universe.Service, txn neo.Starter, market neo.MarketRepository) Service {
	return &service{
		redis,
		esi,
		logger,
		universe,
		txn,
		market,
	}
}
