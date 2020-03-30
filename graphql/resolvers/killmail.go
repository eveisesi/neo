package resolvers

import (
	"context"

	"github.com/eveisesi/neo"
	"github.com/eveisesi/neo/graphql/service"
)

func (r *queryResolver) Killmail(ctx context.Context, id int, hash string) (*neo.Killmail, error) {
	return r.KillmailServ.Killmail(ctx, uint64(id), hash)
}

func (r *queryResolver) KillmailRecent(ctx context.Context, page *int) ([]*neo.Killmail, error) {
	return r.KillmailServ.KillmailRecent(ctx, page)
}

func (r *Resolver) Killmail() service.KillmailResolver {
	return &killmailResolver{r}
}

type killmailResolver struct{ *Resolver }

func (r *killmailResolver) Attackers(ctx context.Context, obj *neo.Killmail) ([]*neo.KillmailAttacker, error) {
	return r.Dataloader(ctx).KillmailAttackersLoader.Load(obj.ID)
}
func (r *killmailResolver) Victim(ctx context.Context, obj *neo.Killmail) (*neo.KillmailVictim, error) {
	return r.Dataloader(ctx).KillmailVictimLoader.Load(obj.ID)
}

func (r *killmailResolver) System(ctx context.Context, obj *neo.Killmail) (*neo.SolarSystem, error) {
	return r.Dataloader(ctx).SolarSystemLoader.Load(obj.SolarSystemID)
}

func (r *Resolver) KillmailAttacker() service.KillmailAttackerResolver {
	return &killmailAttackerResolver{r}
}

type killmailAttackerResolver struct{ *Resolver }

func (r *killmailAttackerResolver) Alliance(ctx context.Context, obj *neo.KillmailAttacker) (*neo.Alliance, error) {
	return r.Dataloader(ctx).AllianceLoader.Load(obj.AllianceID.Uint64)
}

func (r *killmailAttackerResolver) Corporation(ctx context.Context, obj *neo.KillmailAttacker) (*neo.Corporation, error) {
	return r.Dataloader(ctx).CorporationLoader.Load(obj.CorporationID.Uint64)
}

func (r *killmailAttackerResolver) Character(ctx context.Context, obj *neo.KillmailAttacker) (*neo.Character, error) {
	return r.Dataloader(ctx).CharacterLoader.Load(obj.CharacterID.Uint64)
}

func (r *killmailAttackerResolver) Ship(ctx context.Context, obj *neo.KillmailAttacker) (*neo.Type, error) {
	return r.Dataloader(ctx).TypeLoader.Load(obj.ShipTypeID.Uint64)
}

func (r *killmailAttackerResolver) Weapon(ctx context.Context, obj *neo.KillmailAttacker) (*neo.Type, error) {
	return r.Dataloader(ctx).TypeLoader.Load(obj.WeaponTypeID.Uint64)
}

func (r *Resolver) KillmailItem() service.KillmailItemResolver {
	return &killmailItemResolver{r}
}

type killmailItemResolver struct{ *Resolver }

func (r *killmailItemResolver) Type(ctx context.Context, obj *neo.KillmailItem) (*neo.Type, error) {
	return r.Dataloader(ctx).TypeLoader.Load(obj.ItemTypeID)
}

func (r *killmailItemResolver) Items(ctx context.Context, obj *neo.KillmailItem) ([]*neo.KillmailItem, error) {
	return r.Dataloader(ctx).KillmailItemsLoader.Load(&neo.KillmailItemLoader{
		ID:   obj.ParentID.Uint64,
		Type: neo.ChildKillmailItem,
	})
}

func (r *Resolver) KillmailVictim() service.KillmailVictimResolver {
	return &killmailVictimResolver{r}
}

type killmailVictimResolver struct{ *Resolver }

func (r *killmailVictimResolver) Alliance(ctx context.Context, obj *neo.KillmailVictim) (*neo.Alliance, error) {
	return r.Dataloader(ctx).AllianceLoader.Load(obj.AllianceID.Uint64)
}

func (r *killmailVictimResolver) Corporation(ctx context.Context, obj *neo.KillmailVictim) (*neo.Corporation, error) {
	return r.Dataloader(ctx).CorporationLoader.Load(obj.CorporationID)
}

func (r *killmailVictimResolver) Character(ctx context.Context, obj *neo.KillmailVictim) (*neo.Character, error) {
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

func (r *killmailVictimResolver) Items(ctx context.Context, obj *neo.KillmailVictim) ([]*neo.KillmailItem, error) {
	return r.Dataloader(ctx).KillmailItemsLoader.Load(&neo.KillmailItemLoader{
		ID:   obj.KillmailID,
		Type: neo.ParentKillmailItem,
	})
}
