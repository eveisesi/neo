package resolvers

import (
	"context"

	"github.com/eveisesi/neo"
	"github.com/eveisesi/neo/graphql/service"
)

func (r *queryResolver) KillmailsByShipID(ctx context.Context, id int, page *int) ([]*neo.Killmail, error) {
	return r.Services.KillmailsByShipID(ctx, uint64(id), *page)
}

func (r *Resolver) Constellation() service.ConstellationResolver {
	return &constellationResolver{r}
}

type constellationResolver struct{ *Resolver }

func (r *constellationResolver) Region(ctx context.Context, obj *neo.Constellation) (*neo.Region, error) {
	return r.Dataloader(ctx).RegionLoader.Load(obj.RegionID)
}

func (r *Resolver) SolarSystem() service.SolarSystemResolver {
	return &solarSystemResolver{r}
}

type solarSystemResolver struct{ *Resolver }

func (r *solarSystemResolver) Constellation(ctx context.Context, obj *neo.SolarSystem) (*neo.Constellation, error) {
	return r.Dataloader(ctx).ConstellationLoader.Load(obj.ConstellationID)
}
