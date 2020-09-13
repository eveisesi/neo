package tracker

import (
	"context"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/sirupsen/logrus"
)

type Service interface {
	Run(start, end time.Time)
	Watchman(ctx context.Context)
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
