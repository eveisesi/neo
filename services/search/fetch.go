package search

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/RediSearch/redisearch-go/redisearch"
	"github.com/eveisesi/neo"
	"github.com/pkg/errors"
)

func (s *service) Fetch(ctx context.Context, term string) ([]*neo.SearchableEntity, error) {

	fmt.Println(term)

	suggestions, err := s.Autocompleter.SuggestOpts(term, redisearch.SuggestOptions{
		Num:          20,
		Fuzzy:        false,
		WithPayloads: true,
		WithScores:   false,
	})
	if err != nil {
		msg := "failed to fetch suggestions from autocompleter"
		s.Logger.WithError(err).Error(msg)
		return nil, errors.Wrap(err, msg)
	}

	var entities = make([]*neo.SearchableEntity, 0)
	for _, suggestion := range suggestions {
		entity := neo.SearchableEntity{}
		err := json.Unmarshal([]byte(suggestion.Payload), &entity)
		if err != nil {
			msg := "failed to unmarhal suggestion payload"
			s.Logger.WithError(err).WithField("data", suggestion.Payload).Error(msg)
			return nil, errors.Wrap(err, msg)
		}

		entities = append(entities, &entity)
	}

	return entities, nil

}
