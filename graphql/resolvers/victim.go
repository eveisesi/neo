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
	if !obj.AllianceID.Valid {
		return nil, nil
	}
	return r.Dataloader(ctx).AllianceLoader.Load(obj.AllianceID.Uint)
}

func (r *killmailVictimResolver) Corporation(ctx context.Context, obj *neo.KillmailVictim) (*neo.Corporation, error) {
	if !obj.CorporationID.Valid {
		return nil, nil
	}
	return r.Dataloader(ctx).CorporationLoader.Load(obj.CorporationID.Uint)
}

func (r *killmailVictimResolver) Character(ctx context.Context, obj *neo.KillmailVictim) (*neo.Character, error) {
	if !obj.CharacterID.Valid {
		return nil, nil
	}
	return r.Dataloader(ctx).CharacterLoader.Load(obj.CharacterID.Uint64)
}

func (r *killmailVictimResolver) Ship(ctx context.Context, obj *neo.KillmailVictim) (*neo.Type, error) {
	return r.Dataloader(ctx).TypeLoader.Load(obj.ShipTypeID)
}

func (r *killmailVictimResolver) Position(ctx context.Context, obj *neo.KillmailVictim) (*neo.KillmailPosition, error) {
	if obj.PosX.Valid && obj.PosY.Valid && obj.PosZ.Valid {
		return &neo.KillmailPosition{
			X: obj.PosX,
			Y: obj.PosY,
			Z: obj.PosZ,
		}, nil
	}

	return nil, nil
}

func (r *killmailVictimResolver) Fitted(ctx context.Context, obj *neo.KillmailVictim) ([]*neo.KillmailItem, error) {
	items, err := r.Dataloader(ctx).KillmailItemsLoader.Load(obj.KillmailID)
	if err != nil {
		return nil, err
	}
	result := make([]*neo.KillmailItem, 0)
	for _, item := range items {
		if item.Flag == 5 {
			continue
		}

		result = append(result, item)
	}

	return result, nil
}

func (r *killmailVictimResolver) Items(ctx context.Context, obj *neo.KillmailVictim) ([]*neo.KillmailItem, error) {

	items, err := r.Dataloader(ctx).KillmailItemsLoader.Load(obj.KillmailID)
	if err != nil {
		return nil, err
	}

	// itemID => slotSection => destroyed/dropped => item
	store := make(map[uint]map[string]map[string]*neo.KillmailItem)
	for _, item := range items {

		action := ""
		if item.QuantityDestroyed.Valid && item.QuantityDestroyed.Uint > 0 {
			action = "destroyed"
		} else if item.QuantityDropped.Valid && item.QuantityDropped.Uint > 0 {
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
			store[item.ItemTypeID][slot][action].QuantityDestroyed.Uint += item.QuantityDestroyed.Uint
			store[item.ItemTypeID][slot][action].QuantityDropped.Uint += item.QuantityDropped.Uint
		}

	}

	items = make([]*neo.KillmailItem, 0)
	for _, item := range store {
		for _, slot := range item {
			for _, slotItem := range slot {
				slotItem.TotalValue = float64(slotItem.QuantityDropped.Uint+slotItem.QuantityDestroyed.Uint) * slotItem.ItemValue
				items = append(items, slotItem)
			}
		}
	}

	return items, nil
}
