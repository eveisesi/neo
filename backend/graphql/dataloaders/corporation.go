package dataloaders

import (
	"context"

	"github.com/eveisesi/neo"
	"github.com/eveisesi/neo/graphql/dataloaders/generated"
	"github.com/eveisesi/neo/services/corporation"
)

func CorporationLoader(ctx context.Context, corporation corporation.Service) *generated.CorporationLoader {
	return generated.NewCorporationLoader(generated.CorporationLoaderConfig{
		Wait:     defaultWait,
		MaxBatch: defaultMaxBatch,
		Fetch: func(ids []uint) ([]*neo.Corporation, []error) {
			errors := make([]error, len(ids))

			rows, err := corporation.CorporationsByCorporationIDs(ctx, ids)
			if err != nil {
				errors = append(errors, err)
				return nil, errors
			}

			corporationByCorporationID := map[uint]*neo.Corporation{}
			for _, c := range rows {
				corporationByCorporationID[c.ID] = c
			}

			corporations := make([]*neo.Corporation, len(ids))
			for i, x := range ids {
				corporations[i] = corporationByCorporationID[x]
			}

			return corporations, nil
		},
	})
}
