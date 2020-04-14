package resolvers

import (
	"context"

	"github.com/davecgh/go-spew/spew"
	"github.com/eveisesi/neo"
	"github.com/eveisesi/neo/graphql/service"
)

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

func (r *solarSystemResolver) Region(ctx context.Context, obj *neo.SolarSystem) (*neo.Region, error) {
	spew.Dump(obj)
	return r.Dataloader(ctx).RegionLoader.Load(obj.RegionID)
}

func (r *solarSystemResolver) Constellation(ctx context.Context, obj *neo.SolarSystem) (*neo.Constellation, error) {
	return r.Dataloader(ctx).ConstellationLoader.Load(obj.ConstellationID)
}
