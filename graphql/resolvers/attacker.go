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
	if !obj.AllianceID.Valid {
		return nil, nil
	}
	return r.Dataloader(ctx).AllianceLoader.Load(obj.AllianceID.Uint64)
}

func (r *killmailAttackerResolver) Corporation(ctx context.Context, obj *neo.KillmailAttacker) (*neo.Corporation, error) {
	if !obj.CorporationID.Valid {
		return nil, nil
	}
	return r.Dataloader(ctx).CorporationLoader.Load(obj.CorporationID.Uint64)
}

func (r *killmailAttackerResolver) Character(ctx context.Context, obj *neo.KillmailAttacker) (*neo.Character, error) {
	if !obj.CharacterID.Valid {
		return nil, nil
	}
	return r.Dataloader(ctx).CharacterLoader.Load(obj.CharacterID.Uint64)
}

func (r *killmailAttackerResolver) Ship(ctx context.Context, obj *neo.KillmailAttacker) (*neo.Type, error) {
	if !obj.ShipTypeID.Valid {
		return nil, nil
	}
	return r.Dataloader(ctx).TypeLoader.Load(obj.ShipTypeID.Uint64)
}

func (r *killmailAttackerResolver) Weapon(ctx context.Context, obj *neo.KillmailAttacker) (*neo.Type, error) {
	if !obj.WeaponTypeID.Valid {
		return nil, nil
	}
	return r.Dataloader(ctx).TypeLoader.Load(obj.WeaponTypeID.Uint64)
}
