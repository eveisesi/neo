package dataloaders

import (
	"context"

	"github.com/eveisesi/neo"
	"github.com/eveisesi/neo/graphql/dataloaders/generated"
	"github.com/eveisesi/neo/services/universe"
)

func SolarSystemLoader(ctx context.Context, universe universe.Service) *generated.SolarSystemLoader {
	return generated.NewSolarSystemLoader(generated.SolarSystemLoaderConfig{
		Wait:     defaultWait,
		MaxBatch: defaultMaxBatch,
		Fetch: func(ids []uint64) ([]*neo.SolarSystem, []error) {
			solarSystems := make([]*neo.SolarSystem, len(ids))
			errors := make([]error, len(ids))

			rows, err := universe.SolarSystemsBySolarSystemIDs(ctx, ids)
			if err != nil {
				errors = append(errors, err)
				return nil, errors
			}

			solarSystemsBySolarSystemID := map[uint64]*neo.SolarSystem{}
			for _, row := range rows {
				solarSystemsBySolarSystemID[row.ID] = row
			}

			for i, v := range ids {
				solarSystems[i] = solarSystemsBySolarSystemID[v]
			}

			return solarSystems, nil
		},
	})
}

func TypeLoader(ctx context.Context, univsere universe.Service) *generated.TypeLoader {
	return generated.NewTypeLoader(generated.TypeLoaderConfig{
		Wait:     defaultWait,
		MaxBatch: defaultMaxBatch,
		Fetch: func(ids []uint64) ([]*neo.Type, []error) {
			invTypes := make([]*neo.Type, len(ids))
			errors := make([]error, len(ids))

			rows, err := univsere.TypesByTypeIDs(ctx, ids)
			if err != nil {
				errors = append(errors, err)
				return nil, errors
			}

			invTypesByTypeID := map[uint64]*neo.Type{}
			for _, row := range rows {
				invTypesByTypeID[row.ID] = row
			}

			for i, v := range ids {
				invTypes[i] = invTypesByTypeID[v]
			}

			return invTypes, nil
		},
	})
}
