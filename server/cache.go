package server

import (
	"context"
	"fmt"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/eveisesi/neo"
	"github.com/go-redis/redis/v7"
)

type GQLCache struct {
	client *redis.Client
	ttl    time.Duration
}

func (c *GQLCache) Add(ctx context.Context, hash string, query interface{}) {
	c.client.Set(fmt.Sprintf("%s:%s", neo.REDIS_GRAPHQL_APQ_CACHE, hash), query, c.ttl)
}

func (c *GQLCache) Get(ctx context.Context, hash string) (interface{}, bool) {

	s, err := c.client.Get(hash).Result()
	if err != nil {
		return "", false
	}

	spew.Dump(s)

	return s, true
}
