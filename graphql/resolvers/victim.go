package resolvers

import (
	"context"

	"github.com/eveisesi/neo/tools"

	"github.com/eveisesi/neo"
	"github.com/eveisesi/neo/graphql/service"
)

func (r *Resolver) KillmailVictim() service.KillmailVictimResolver {
	return &killmailVictimResolver{r}
}

type killmailVictimResolver struct{ *Resolver }

func (r *killmailVictimResolver) Alliance(ctx context.Context, obj *neo.KillmailVictim) (*neo.Alliance, error) {
	if obj.AllianceID == nil {
		return nil, nil
	}
	return r.Dataloader(ctx).AllianceLoader.Load(*obj.AllianceID)
}

func (r *killmailVictimResolver) Corporation(ctx context.Context, obj *neo.KillmailVictim) (*neo.Corporation, error) {
	if obj.CorporationID == nil {
		return nil, nil
	}
	return r.Dataloader(ctx).CorporationLoader.Load(*obj.CorporationID)
}

func (r *killmailVictimResolver) Character(ctx context.Context, obj *neo.KillmailVictim) (*neo.Character, error) {
	if obj.CharacterID == nil {
		return nil, nil
	}
	return r.Dataloader(ctx).CharacterLoader.Load(*obj.CharacterID)
}

func (r *killmailVictimResolver) Ship(ctx context.Context, obj *neo.KillmailVictim) (*neo.Type, error) {
	return r.Dataloader(ctx).TypeLoader.Load(obj.ShipTypeID)
}

func (r *killmailVictimResolver) Fitted(ctx context.Context, obj *neo.KillmailVictim) ([]*neo.KillmailItem, error) {

	if len(obj.Items) == 0 {
		return obj.Items, nil
	}

	result := make([]*neo.KillmailItem, 0)
	for _, item := range obj.Items {
		if item.Flag == 5 {
			continue
		}

		result = append(result, item)
	}

	return result, nil
}

func (r *killmailVictimResolver) Items(ctx context.Context, obj *neo.KillmailVictim) ([]*neo.KillmailItem, error) {

	if len(obj.Items) == 0 {
		return obj.Items, nil
	}

	items := obj.Items

	// itemID => slotSection => destroyed/dropped => item
	store := make(map[uint]map[string]map[string]*neo.KillmailItem)
	for _, item := range items {

		action := ""
		if item.QuantityDestroyed != nil && *item.QuantityDestroyed > 0 {
			action = "destroyed"
		} else if item.QuantityDropped != nil && *item.QuantityDropped > 0 {
			action = "dropped"
		}

		slot := tools.SlotForFlagID(item.Flag)

		if _, ok := store[item.ItemTypeID]; !ok {
			store[item.ItemTypeID] = make(map[string]map[string]*neo.KillmailItem)
		}
		if _, ok := store[item.ItemTypeID][slot]; !ok {
			store[item.ItemTypeID][slot] = make(map[string]*neo.KillmailItem)
		}

		if store[item.ItemTypeID][slot][action] == nil {
			store[item.ItemTypeID][slot][action] = item
		} else {
			pos := store[item.ItemTypeID][slot][action]
			if pos.QuantityDestroyed == nil && item.QuantityDestroyed != nil && *item.QuantityDestroyed > 0 {
				*pos.QuantityDestroyed = *item.QuantityDestroyed
			}

			if pos.QuantityDropped == nil && item.QuantityDropped != nil && *item.QuantityDropped > 0 {
				*pos.QuantityDropped = *item.QuantityDropped
			}

			// .QuantityDestroyed += item.QuantityDestroyed
			// store[item.ItemTypeID][slot][action].QuantityDropped += item.QuantityDropped
		}

	}

	items = make([]*neo.KillmailItem, 0)
	for _, item := range store {
		for _, slot := range item {
			for _, slotItem := range slot {
				totalValue := float64(0)
				if slotItem.QuantityDropped != nil && *slotItem.QuantityDropped > 0 {
					totalValue += float64(*slotItem.QuantityDropped) * slotItem.ItemValue
				}
				if slotItem.QuantityDestroyed != nil && *slotItem.QuantityDestroyed > 0 {
					totalValue += float64(*slotItem.QuantityDestroyed) * slotItem.ItemValue
				}
				slotItem.TotalValue = totalValue
				items = append(items, slotItem)
			}
		}
	}

	return items, nil
}
