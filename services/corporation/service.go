package corporation

import (
	"github.com/eveisesi/neo"
	"github.com/eveisesi/neo/esi"
	"github.com/go-redis/redis"
)

type Service interface {
	neo.CorporationRespository
}

type service struct {
	redis *redis.Client
	esi   *esi.Client
	neo.CorporationRespository
}

func NewService(redis *redis.Client, esi *esi.Client, corporation neo.CorporationRespository) Service {
	return &service{
		redis,
		esi,
		corporation,
	}
}
