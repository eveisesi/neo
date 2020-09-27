package search

import (
	"context"

	"github.com/RediSearch/redisearch-go/redisearch"
	"github.com/eveisesi/neo"
	"github.com/sirupsen/logrus"
)

type Service interface {
	Build(ctx context.Context) error
	Fetch(ctx context.Context, term string) ([]neo.SearchableEntity, error)
	SearchableEntities(ctx context.Context) ([]neo.SearchableEntity, error)
}

type service struct {
	autocompleter *redisearch.Autocompleter
	logger        *logrus.Logger
	character     neo.CharacterRespository
	corporation   neo.CorporationRespository
	alliance      neo.AllianceRespository
	universe      neo.UniverseRepository
}

func NewService(
	autocompleter *redisearch.Autocompleter,
	logger *logrus.Logger,
	character neo.CharacterRespository,
	corporation neo.CorporationRespository,
	alliance neo.AllianceRespository,
	universe neo.UniverseRepository,
) Service {

	return &service{
		autocompleter,
		logger,
		character,
		corporation,
		alliance,
		universe,
	}

}
