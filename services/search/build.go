package search

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/eveisesi/neo"

	"github.com/RediSearch/redisearch-go/redisearch"
	"github.com/pkg/errors"
)

func (s *service) Build(ctx context.Context) error {

	// TODO: Wrap this in a DatastoreSegment
	err := s.autocompleter.Delete()
	if err != nil {
		s.logger.WithError(err).Error("failed to flush autocompleter")
		return err
	}

	entities, err := s.SearchableEntities(ctx)
	if err != nil {
		return err
	}

	suggestions := make([]redisearch.Suggestion, 0)

	count := 0

	for _, entity := range entities {
		payload, err := json.Marshal(entity)
		if err != nil {
			s.logger.WithContext(ctx).WithError(err).Error("failed to marshal searchable entity")
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
			err = s.autocompleter.AddTerms(suggestions...)
			if err != nil {
				msg := "failed to add searchable entities to autocompleter"
				s.logger.WithContext(ctx).WithError(err).Error(msg)
				return errors.Wrap(err, msg)
			}

			count = 0
			suggestions = make([]redisearch.Suggestion, 0)

		}

	}

	if len(suggestions) > 0 {
		// Finishing up the last of the suggestions
		err = s.autocompleter.AddTerms(suggestions...)
		if err != nil {
			msg := "failed to add searchable entities to autocompleter"
			s.logger.WithContext(ctx).WithError(err).Error(msg)
			return errors.Wrap(err, msg)
		}

	}

	return err

}

func (s *service) SearchableEntities(ctx context.Context) ([]neo.SearchableEntity, error) {

	entities := make([]neo.SearchableEntity, 0)
	characters, err := s.character.Characters(ctx, []neo.Modifier{}...)
	if err != nil {
		s.logger.WithError(err).Error("failed to fetch characters")
		return nil, errors.New("failed to fetch characters")
	}

	for _, character := range characters {
		entities = append(entities, neo.SearchableEntity{
			ID:       character.ID,
			Name:     character.Name,
			Type:     "characters",
			Image:    fmt.Sprintf("characters/%d/portrait", character.ID),
			Priority: 1,
		})
	}

	corporations, err := s.corporation.Corporations(ctx, []neo.Modifier{}...)
	if err != nil {
		s.logger.WithError(err).Error("failed to fetch corporations")
		return nil, errors.New("failed to fetch corporations")
	}

	for _, corporation := range corporations {
		entities = append(entities, neo.SearchableEntity{
			ID:       uint64(corporation.ID),
			Name:     corporation.Name,
			Type:     "corporations",
			Image:    fmt.Sprintf("corporations/%d/logo", corporation.ID),
			Priority: 1,
		})
	}

	alliances, err := s.alliance.Alliances(ctx, []neo.Modifier{}...)
	if err != nil {
		s.logger.WithError(err).Error("failed to fetch alliances")
		return nil, errors.New("failed to fetch alliances")
	}

	for _, alliance := range alliances {
		entities = append(entities, neo.SearchableEntity{
			ID:       uint64(alliance.ID),
			Name:     alliance.Name,
			Type:     "alliances",
			Image:    fmt.Sprintf("alliances/%d/logo", alliance.ID),
			Priority: 1,
		})
	}

	groupIDs := make([]neo.ModValue, 0)
	shipGroups, err := s.universe.TypeGroups(ctx, neo.EqualTo{Column: "categoryID", Value: 6})
	if err != nil {
		s.logger.WithError(err).Error("failed to fetch ship groups")
		return nil, errors.New("failed to fetch ship groups")
	}

	for _, group := range shipGroups {
		groupIDs = append(groupIDs, group.ID)
	}

	ships, err := s.universe.Types(ctx, neo.In{Column: "groupID", Values: groupIDs})
	if err != nil {
		s.logger.WithError(err).Error("failed to fetch ships")
		return nil, errors.New("failed to fetch ships")
	}

	for _, ship := range ships {
		entities = append(entities, neo.SearchableEntity{
			ID:       uint64(ship.ID),
			Name:     ship.Name,
			Type:     "types",
			Image:    fmt.Sprintf("ships/%d/render", ship.ID),
			Priority: 2,
		})
	}

	solarSystems, err := s.universe.SolarSystems(ctx, []neo.Modifier{}...)
	if err != nil {
		s.logger.WithError(err).Error("failed to fetch solar systems")
		return nil, errors.New("failed to fetch solar systems")
	}

	for _, system := range solarSystems {
		entities = append(entities, neo.SearchableEntity{
			ID:       uint64(system.ID),
			Name:     system.Name,
			Type:     "systems",
			Image:    "types/6/render",
			Priority: 3,
		})
	}

	constellations, err := s.universe.Constellations(ctx, []neo.Modifier{}...)
	if err != nil {
		s.logger.WithError(err).Error("failed to fetch constellations")
		return nil, errors.New("failed to fetch constellations")
	}

	for _, constellation := range constellations {
		entities = append(entities, neo.SearchableEntity{
			ID:       uint64(constellation.ID),
			Name:     constellation.Name,
			Type:     "constellations",
			Image:    "types/7/render",
			Priority: 3,
		})
	}

	regions, err := s.universe.Regions(ctx, []neo.Modifier{}...)
	if err != nil {
		s.logger.WithError(err).Error("failed to fetch regions")
		return nil, errors.New("failed to fetch regions")
	}

	for _, region := range regions {
		entities = append(entities, neo.SearchableEntity{
			ID:       uint64(region.ID),
			Name:     region.Name,
			Type:     "regions",
			Image:    "types/8/render",
			Priority: 3,
		})
	}

	return entities, err

}
