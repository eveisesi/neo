package resolvers

import (
	"context"

	"github.com/eveisesi/neo"
	"github.com/eveisesi/neo/graphql/service"
)

func (r *Resolver) Type() service.TypeResolver {
	return &typeResolver{r}
}

type typeResolver struct {
	*Resolver
}

func (r *typeResolver) Attributes(ctx context.Context, obj *neo.Type) ([]*neo.TypeAttribute, error) {
	return r.Dataloader(ctx).TypeAttributeLoader.Load(obj.ID)
}

func (r *typeResolver) Group(ctx context.Context, obj *neo.Type) (*neo.TypeGroup, error) {
	return r.Dataloader(ctx).TypeGroupLoader.Load(obj.GroupID)
}

func (r *Resolver) TypeGroup() service.TypeGroupResolver {
	return &typeGroupResolver{r}
}

type typeGroupResolver struct {
	*Resolver
}

func (r *typeGroupResolver) Category(ctx context.Context, obj *neo.TypeGroup) (*neo.TypeCategory, error) {
	return r.Dataloader(ctx).TypeCategoryLoader.Load(obj.CategoryID)
}
