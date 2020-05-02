package resolvers

import (
	"context"

	"github.com/volatiletech/null"

	"github.com/eveisesi/neo"
	"github.com/eveisesi/neo/graphql/service"
)

func (r *queryResolver) Killmail(ctx context.Context, id int, hash string) (*neo.Killmail, error) {
	return r.Services.Killmail.Killmail(ctx, uint64(id), hash)
}

func (r *queryResolver) KillmailRecent(ctx context.Context, page *int) ([]*neo.Killmail, error) {
	return r.Services.Killmail.KillmailRecent(ctx, null.IntFromPtr(page))
}

func (r *queryResolver) KillmailTopByAge(ctx context.Context, age *int, limit *int) ([]*neo.Killmail, error) {

	newAge := *age
	newLimit := *limit

	return r.Services.KillmailTop(ctx, uint64(newAge), uint64(newLimit))

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
