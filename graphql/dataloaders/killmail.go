package dataloaders

import (
	"context"

	"github.com/eveisesi/neo"
	"github.com/eveisesi/neo/graphql/dataloaders/generated"
	"github.com/eveisesi/neo/services/killmail"
)

func KillmailAttackersLoader(ctx context.Context, killmail killmail.Service) *generated.KillmailAttackersLoader {
	return generated.NewKillmailAttackersLoader(generated.KillmailAttackersLoaderConfig{
		Wait:     defaultWait,
		MaxBatch: defaultMaxBatch,
		Fetch: func(ids []uint64) ([][]*neo.KillmailAttacker, []error) {

			attackers := make([][]*neo.KillmailAttacker, len(ids))
			errors := make([]error, len(ids))

			rows, err := killmail.AttackersByKillmailIDs(ctx, ids)
			if err != nil {
				errors = append(errors, err)
				return nil, errors
			}

			attackersByKillmailID := map[uint64][]*neo.KillmailAttacker{}
			for _, row := range rows {
				attackersByKillmailID[row.KillmailID] = append(attackersByKillmailID[row.KillmailID], row)
			}

			for i, v := range ids {
				attackers[i] = attackersByKillmailID[v]
			}

			return attackers, nil

		},
	})
}

func KillmailItemsLoader(ctx context.Context, killmail killmail.Service) *generated.KillmailItemsLoader {
	return generated.NewKillmailItemsLoader(generated.KillmailItemsLoaderConfig{
		Wait:     defaultWait,
		MaxBatch: defaultMaxBatch,
		Fetch: func(ids []uint64) ([][]*neo.KillmailItem, []error) {

			items := make([][]*neo.KillmailItem, len(ids))
			errors := make([]error, len(ids))

			rows, err := killmail.ItemsByKillmailIDs(ctx, ids)
			if err != nil {
				errors = append(errors, err)
				return nil, errors
			}

			var parentsWithChildren = make([]*neo.KillmailItem, 0)
			for _, row := range rows {
				if !row.ParentID.Valid {
					parentsWithChildren = append(parentsWithChildren, row)
				}
				if row.ParentID.Valid {
					for i, parent := range parentsWithChildren {
						if parent.ID == row.ParentID.Uint64 {
							parentsWithChildren[i].Items = append(parentsWithChildren[i].Items, row)
							break
						}
					}
				}
			}

			var itemsByKillmailID = make(map[uint64][]*neo.KillmailItem)
			for _, row := range parentsWithChildren {
				itemsByKillmailID[row.KillmailID] = append(itemsByKillmailID[row.KillmailID], row)
			}

			for i, v := range ids {
				items[i] = itemsByKillmailID[v]
			}

			return items, nil
		},
	})
}

func KillmailVictimLoader(ctx context.Context, killmail killmail.Service) *generated.KillmailVictimLoader {
	return generated.NewKillmailVictimLoader(generated.KillmailVictimLoaderConfig{
		Wait:     defaultWait,
		MaxBatch: defaultMaxBatch,
		Fetch: func(ids []uint64) ([]*neo.KillmailVictim, []error) {
			victims := make([]*neo.KillmailVictim, len(ids))
			errors := make([]error, len(ids))

			rows, err := killmail.VictimsByKillmailIDs(ctx, ids)
			if err != nil {
				errors = append(errors, err)
				return nil, errors
			}

			victimByKillmailID := map[uint64]*neo.KillmailVictim{}
			for _, row := range rows {
				victimByKillmailID[row.KillmailID] = row
			}

			for i, v := range ids {
				victims[i] = victimByKillmailID[v]
			}

			return victims, nil
		},
	})
}
