package resolvers

import (
	"context"

	"github.com/eveisesi/neo/graphql/dataloaders"
	"github.com/eveisesi/neo/graphql/service"
	"github.com/eveisesi/neo/services/killmail"
)

// THIS CODE IS A STARTING POINT ONLY. IT WILL NOT BE UPDATED WITH SCHEMA CHANGES.

type Resolver struct {
	Services   Services
	Dataloader func(ctx context.Context) dataloaders.Loaders
}

type Killmail killmail.Service

type Services struct {
	Killmail
}

func (r *Resolver) Mutation() service.MutationResolver {
	return &mutationResolver{r}
}

func (r *Resolver) Query() service.QueryResolver {
	return &queryResolver{r}
}

type mutationResolver struct{ *Resolver }

func (r *mutationResolver) MutationPlaceholder(ctx context.Context) (bool, error) {
	panic("not implemented")
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) QueryPlaceholder(ctx context.Context) (bool, error) {
	panic("not implemented")
}
