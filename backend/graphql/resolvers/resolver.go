package resolvers

import (
	"context"

	"github.com/eveisesi/neo"
	"github.com/eveisesi/neo/services/alliance"
	"github.com/eveisesi/neo/services/character"
	"github.com/eveisesi/neo/services/corporation"
	"github.com/eveisesi/neo/services/search"
	"github.com/eveisesi/neo/services/universe"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"

	"github.com/eveisesi/neo/graphql/dataloaders"
	"github.com/eveisesi/neo/graphql/service"
	"github.com/eveisesi/neo/services/killmail"
)

// THIS CODE IS A STARTING POINT ONLY. IT WILL NOT BE UPDATED WITH SCHEMA CHANGES.

type Resolver struct {
	Services   Services
	Dataloader func(ctx context.Context) dataloaders.Loaders
	Logger     *logrus.Logger
	Redis      *redis.Client
}

type Killmail killmail.Service
type Alliance alliance.Service
type Corporation corporation.Service
type Character character.Service
type Universe universe.Service
type Search search.Service

type Services struct {
	Killmail
	Alliance
	Corporation
	Character
	Universe
	Search
}

type FeedManager struct {
	TotalConnections int
	Feeds            []struct {
		Receiver        chan *neo.Killmail
		ConnectionCount int
		Subscribers     map[string](chan *neo.Killmail)
	}
}

func NewResolver(
	ctxLoaders func(ctx context.Context) dataloaders.Loaders,
	logger *logrus.Logger,
	redis *redis.Client,
	services Services,
) *Resolver {

	redis.Del(context.Background(), neo.WEBSOCKET_CONNECTIONS)
	redis.Set(context.Background(), neo.WEBSOCKET_CONNECTIONS, 0, -1)

	return &Resolver{
		Services:   services,
		Dataloader: ctxLoaders,
		Logger:     logger,
		Redis:      redis,
	}

}

var (
	err error
)

func (r *Resolver) Mutation() service.MutationResolver {
	return &mutationResolver{r}
}

func (r *Resolver) Query() service.QueryResolver {
	return &queryResolver{r}
}

func (r *Resolver) Subscription() service.SubscriptionResolver {
	return &subscriptionResolver{r}
}

type subscriptionResolver struct{ *Resolver }

type mutationResolver struct{ *Resolver }

func (r *mutationResolver) MutationPlaceholder(ctx context.Context) (bool, error) {
	panic("not implemented")
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) QueryPlaceholder(ctx context.Context) (bool, error) {
	panic("not implemented")
}
