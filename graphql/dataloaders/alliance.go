package dataloaders

import (
	"context"

	"github.com/ddouglas/killboard"
	"github.com/ddouglas/killboard/graphql/dataloaders/generated"
	"github.com/ddouglas/killboard/services/alliance"
)

func AllianceLoader(ctx context.Context, alliance alliance.Service) *generated.AllianceLoader {
	return generated.NewAllianceLoader(generated.AllianceLoaderConfig{
		Wait:     defaultWait,
		MaxBatch: defaultMaxBatch,
		Fetch: func(ids []uint64) ([]*killboard.Alliance, []error) {

			alliances := make([]*killboard.Alliance, len(ids))
			errors := make([]error, len(ids))

			rows, err := alliance.AlliancesByAllianceIDs(ctx, ids)
			if err != nil {
				errors = append(errors, err)
				return nil, errors
			}

			allianceByAllianceID := map[uint64]*killboard.Alliance{}
			for _, c := range rows {
				allianceByAllianceID[c.ID] = c
			}

			for i, x := range ids {
				alliances[i] = allianceByAllianceID[x]
			}

			return alliances, nil

		},
	})
}
