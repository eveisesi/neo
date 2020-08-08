package search

import (
	"context"
	"encoding/json"

	"github.com/RediSearch/redisearch-go/redisearch"
	"github.com/pkg/errors"
)

func (s *service) Build(ctx context.Context) error {

	// TODO: Wrap this in a DatastoreSegment
	err := s.Autocompleter.Delete()
	if err != nil {
		s.Logger.WithError(err).Error("failed to flush autocompleter")
		return err
	}

	entities, err := s.AllSearchableEntities(ctx)
	if err != nil {
		return err
	}

	suggestions := make([]redisearch.Suggestion, 0)

	count := 0

	for _, entity := range entities {
		payload, err := json.Marshal(entity)
		if err != nil {
			s.Logger.WithContext(ctx).WithError(err).Error("failed to marshal searchable entity")
			continue
		}

		suggestion := redisearch.Suggestion{
			Term:    entity.Name,
			Score:   float64(entity.Priority),
			Payload: string(payload),
		}

		suggestions = append(suggestions, suggestion)

		// We don't want to hammer redis, so exec an add every 5K terms and then reset the slice
		count++
		if count >= 5000 {
			// Wrap in DatastoreSegment
			err = s.Autocompleter.AddTerms(suggestions...)
			if err != nil {
				msg := "failed to add searchable entities to autocompleter"
				s.Logger.WithContext(ctx).WithError(err).Error(msg)
				return errors.Wrap(err, msg)
			}

			count = 0
			suggestions = make([]redisearch.Suggestion, 0)

		}

	}

	if len(suggestions) > 0 {
		// Finishing up the last of the suggestions
		err = s.Autocompleter.AddTerms(suggestions...)
		if err != nil {
			msg := "failed to add searchable entities to autocompleter"
			s.Logger.WithContext(ctx).WithError(err).Error(msg)
			return errors.Wrap(err, msg)
		}

	}

	return err

}
