package tracker

import (
	"time"

	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
)

type Service interface {
	Run(start, end time.Time)
	GateKeeper()
}

type service struct {
	redis  *redis.Client
	logger *logrus.Logger
}

func NewService(redis *redis.Client, logger *logrus.Logger) Service {
	return &service{
		redis,
		logger,
	}
}
