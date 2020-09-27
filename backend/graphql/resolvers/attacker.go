package resolvers

import (
	"context"

	"github.com/eveisesi/neo"
	"github.com/eveisesi/neo/graphql/service"
)

func (r *Resolver) KillmailAttacker() service.KillmailAttackerResolver {
	return &killmailAttackerResolver{r}
}

type killmailAttackerResolver struct{ *Resolver }

func (r *killmailAttackerResolver) Alliance(ctx context.Context, obj *neo.KillmailAttacker) (*neo.Alliance, error) {
	if obj.AllianceID == nil {
		return nil, nil
	}
	return r.Dataloader(ctx).AllianceLoader.Load(*obj.AllianceID)
}

func (r *killmailAttackerResolver) Corporation(ctx context.Context, obj *neo.KillmailAttacker) (*neo.Corporation, error) {
	if obj.CorporationID == nil {
		return nil, nil
	}
	return r.Dataloader(ctx).CorporationLoader.Load(*obj.CorporationID)
}

func (r *killmailAttackerResolver) Character(ctx context.Context, obj *neo.KillmailAttacker) (*neo.Character, error) {
	if obj.CharacterID == nil {
		return nil, nil
	}
	return r.Dataloader(ctx).CharacterLoader.Load(*obj.CharacterID)
}

func (r *killmailAttackerResolver) Ship(ctx context.Context, obj *neo.KillmailAttacker) (*neo.Type, error) {
	if obj.ShipTypeID == nil {
		return nil, nil
	}
	return r.Dataloader(ctx).TypeLoader.Load(*obj.ShipTypeID)
}

func (r *killmailAttackerResolver) Weapon(ctx context.Context, obj *neo.KillmailAttacker) (*neo.Type, error) {
	if obj.WeaponTypeID == nil {
		return nil, nil
	}
	return r.Dataloader(ctx).TypeLoader.Load(*obj.WeaponTypeID)
}
