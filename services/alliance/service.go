package alliance

import (
	"context"

	"github.com/eveisesi/neo"
	"github.com/eveisesi/neo/services/esi"
	"github.com/go-redis/redis"
)

type Service interface {
	Alliance(ctx context.Context, id uint64) (*neo.Alliance, error)
	AlliancesByAllianceIDs(ctx context.Context, ids []uint64) ([]*neo.Alliance, error)
}

type service struct {
	redis *redis.Client
	esi   esi.Service
	neo.AllianceRespository
}

func NewService(redis *redis.Client, esi esi.Service, alliance neo.AllianceRespository) Service {
	return &service{
		redis,
		esi,
		alliance,
	}
}
