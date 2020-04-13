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

func TypeLoader(ctx context.Context, universe universe.Service) *generated.TypeLoader {
	return generated.NewTypeLoader(generated.TypeLoaderConfig{
		Wait:     defaultWait,
		MaxBatch: defaultMaxBatch,
		Fetch: func(ids []uint64) ([]*neo.Type, []error) {
			invTypes := make([]*neo.Type, len(ids))
			errors := make([]error, len(ids))

			rows, err := universe.TypesByTypeIDs(ctx, ids)
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

func TypeAttributeLoader(ctx context.Context, universe universe.Service) *generated.TypeAttributeLoader {
	return generated.NewTypeAttributeLoader(generated.TypeAttributeLoaderConfig{
		Wait:     defaultWait,
		MaxBatch: defaultMaxBatch,
		Fetch: func(ids []uint64) ([][]*neo.TypeAttribute, []error) {
			invTypeAttributes := make([][]*neo.TypeAttribute, len(ids))
			errors := make([]error, len(ids))

			rows, err := universe.TypeAttributesByTypeIDs(ctx, ids)
			if err != nil {
				errors = append(errors, err)
				return nil, errors
			}

			invTypeAttributesByTypeAttributeID := map[uint64][]*neo.TypeAttribute{}
			for _, row := range rows {
				invTypeAttributesByTypeAttributeID[row.TypeID] = append(invTypeAttributesByTypeAttributeID[row.TypeID], row)
			}

			for i, v := range ids {
				invTypeAttributes[i] = invTypeAttributesByTypeAttributeID[v]
			}

			return invTypeAttributes, nil
		},
	})
}

func TypeCategoryLoader(ctx context.Context, universe universe.Service) *generated.TypeCategoryLoader {
	return generated.NewTypeCategoryLoader(generated.TypeCategoryLoaderConfig{
		Wait:     defaultWait,
		MaxBatch: defaultMaxBatch,
		Fetch: func(ids []uint64) ([]*neo.TypeCategory, []error) {
			invTypeCategories := make([]*neo.TypeCategory, len(ids))
			errors := make([]error, len(ids))

			rows, err := universe.TypeCategoriesByCategoryIDs(ctx, ids)
			if err != nil {
				errors = append(errors, err)
				return nil, errors
			}

			invTypeCategoriesByCategoryID := map[uint64]*neo.TypeCategory{}
			for _, row := range rows {
				invTypeCategoriesByCategoryID[row.ID] = row
			}

			for i, v := range ids {
				invTypeCategories[i] = invTypeCategoriesByCategoryID[v]
			}

			return invTypeCategories, nil
		},
	})
}

func TypeFlagLoader(ctx context.Context, universe universe.Service) *generated.TypeFlagLoader {
	return generated.NewTypeFlagLoader(generated.TypeFlagLoaderConfig{
		Wait:     defaultWait,
		MaxBatch: defaultMaxBatch,
		Fetch: func(ids []uint64) ([]*neo.TypeFlag, []error) {
			invTypeFlags := make([]*neo.TypeFlag, len(ids))
			errors := make([]error, len(ids))

			rows, err := universe.TypeFlagsByTypeFlagIDs(ctx, ids)
			if err != nil {
				errors = append(errors, err)
				return nil, errors
			}

			invTypeFlagsByTypeFlagID := map[uint64]*neo.TypeFlag{}
			for _, row := range rows {
				invTypeFlagsByTypeFlagID[row.ID] = row
			}

			for i, v := range ids {
				invTypeFlags[i] = invTypeFlagsByTypeFlagID[v]
			}

			return invTypeFlags, nil
		},
	})
}

func TypeGroupLoader(ctx context.Context, universe universe.Service) *generated.TypeGroupLoader {
	return generated.NewTypeGroupLoader(generated.TypeGroupLoaderConfig{
		Wait:     defaultWait,
		MaxBatch: defaultMaxBatch,
		Fetch: func(ids []uint64) ([]*neo.TypeGroup, []error) {
			invTypeGroups := make([]*neo.TypeGroup, len(ids))
			errors := make([]error, len(ids))

			rows, err := universe.TypeGroupsByGroupIDs(ctx, ids)
			if err != nil {
				errors = append(errors, err)
				return nil, errors
			}

			invTypeGroupsByGroupID := map[uint64]*neo.TypeGroup{}
			for _, row := range rows {
				invTypeGroupsByGroupID[row.ID] = row
			}

			for i, v := range ids {
				invTypeGroups[i] = invTypeGroupsByGroupID[v]
			}

			return invTypeGroups, nil
		},
	})
}
