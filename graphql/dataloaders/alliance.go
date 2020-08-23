package dataloaders

import (
	"context"

	"github.com/eveisesi/neo"
	"github.com/eveisesi/neo/graphql/dataloaders/generated"
	"github.com/eveisesi/neo/services/alliance"
)

func AllianceLoader(ctx context.Context, alliance alliance.Service) *generated.AllianceLoader {
	return generated.NewAllianceLoader(generated.AllianceLoaderConfig{
		Wait:     defaultWait,
		MaxBatch: defaultMaxBatch,
		Fetch: func(ids []uint) ([]*neo.Alliance, []error) {

			errors := make([]error, len(ids))

			rows, err := alliance.AlliancesByAllianceIDs(ctx, ids)
			if err != nil {
				errors = append(errors, err)
				return nil, errors
			}

			allianceByAllianceID := map[uint]*neo.Alliance{}
			for _, c := range rows {
				allianceByAllianceID[c.ID] = c
			}

			alliances := make([]*neo.Alliance, len(ids))
			for i, x := range ids {
				alliances[i] = allianceByAllianceID[x]
			}

			return alliances, nil

		},
	})
}
