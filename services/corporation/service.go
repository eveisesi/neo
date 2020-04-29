package corporation

import (
	"github.com/eveisesi/neo"
	"github.com/eveisesi/neo/services/esi"
	"github.com/go-redis/redis"
)

type Service interface {
	neo.CorporationRespository
}

type service struct {
	redis *redis.Client
	esi   esi.Service
	neo.CorporationRespository
}

func NewService(redis *redis.Client, esi esi.Service, corporation neo.CorporationRespository) Service {
	return &service{
		redis,
		esi,
		corporation,
	}
}
