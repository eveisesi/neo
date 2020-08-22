package resolvers

import (
	"context"

	"github.com/eveisesi/neo"
	"github.com/eveisesi/neo/graphql/service"
)

func (r *queryResolver) GroupByGroupID(ctx context.Context, id int) (*neo.TypeGroup, error) {
	return r.Services.TypeGroup(ctx, uint(id))
}

func (r *queryResolver) CategoryByGroupID(ctx context.Context, id int) (*neo.TypeCategory, error) {
	return r.Services.TypeCategory(ctx, uint(id))
}

func (r *queryResolver) SolarSystemBySolarSystemID(ctx context.Context, id int) (*neo.SolarSystem, error) {
	return r.Services.SolarSystem(ctx, uint(id))
}

func (r *queryResolver) ConstellationByConstellationID(ctx context.Context, id int) (*neo.Constellation, error) {
	return r.Services.Constellation(ctx, uint(id))
}

func (r *queryResolver) RegionByRegionID(ctx context.Context, id int) (*neo.Region, error) {
	return r.Services.Region(ctx, uint(id))
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
