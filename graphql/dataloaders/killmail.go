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

			rows, err := killmail.KillmailAttackersByKillmailIDs(ctx, ids)
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
		Fetch: func(objs []*neo.KillmailItemLoader) ([][]*neo.KillmailItem, []error) {

			items := make([][]*neo.KillmailItem, len(objs))
			errors := make([]error, len(objs))

			itemsByID := map[uint64][]*neo.KillmailItem{}

			killmailIDs := make([]uint64, 0)
			for _, v := range objs {
				if v.Type.IsValid() && v.Type == neo.ParentKillmailItem {
					killmailIDs = append(killmailIDs, v.ID)
				}
			}
			if len(killmailIDs) > 0 {
				parentRows, err := killmail.KillmailItemsByKillmailIDs(ctx, killmailIDs)
				if err != nil {
					errors = append(errors, err)
					return nil, errors
				}

				for _, row := range parentRows {
					itemsByID[row.KillmailID] = append(itemsByID[row.KillmailID], row)
				}
			}

			parentIDs := make([]uint64, 0)
			for _, v := range objs {
				if v.Type.IsValid() && v.Type == neo.ChildKillmailItem {
					parentIDs = append(parentIDs, v.ID)
				}
			}

			if len(parentIDs) > 0 {
				childRows, err := killmail.KillmailItemsByParentIDs(ctx, parentIDs)
				if err != nil {
					errors = append(errors, err)
					return nil, errors
				}
				for _, row := range childRows {
					itemsByID[row.ParentID.Uint64] = append(itemsByID[row.ParentID.Uint64], row)
				}
			}

			for i, v := range objs {
				items[i] = itemsByID[v.ID]
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

			rows, err := killmail.KillmailVictimsByKillmailIDs(ctx, ids)
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
