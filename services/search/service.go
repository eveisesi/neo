package search

import (
	"context"

	"github.com/RediSearch/redisearch-go/redisearch"
	"github.com/eveisesi/neo"
	"github.com/sirupsen/logrus"
)

type Service interface {
	Build(ctx context.Context) error
	Fetch(ctx context.Context, term string) ([]*neo.SearchableEntity, error)
	neo.SearchRepository
}

type service struct {
	*redisearch.Autocompleter
	*logrus.Logger
	neo.SearchRepository
}

func NewService(autocompleter *redisearch.Autocompleter, logger *logrus.Logger, search neo.SearchRepository) Service {

	return &service{
		autocompleter,
		logger,
		search,
	}

}
