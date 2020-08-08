package app

import (
	"context"

	"github.com/eveisesi/neo"
	"github.com/go-redis/redis/v7"
	newrelic "github.com/newrelic/go-agent/v3/newrelic"
)

type redisHook struct {
	cfg *neo.Config
}

func (r redisHook) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {

	txn := newrelic.FromContext(ctx)
	if txn == nil {
		return ctx, nil
	}

	ds := &newrelic.DatastoreSegment{
		StartTime:    txn.StartSegmentNow(),
		Product:      newrelic.DatastoreRedis,
		Operation:    cmd.Name(),
		PortPathOrID: r.cfg.RedisAddr,
	}

	ds.End()

	return ctx, nil

}

func (r redisHook) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	return nil
}

func (r redisHook) BeforeProcessPipeline(ctx context.Context, cmd []redis.Cmder) (context.Context, error) {
	return ctx, nil
}

func (r redisHook) AfterProcessPipeline(ctx context.Context, cmd []redis.Cmder) error {
	return nil
}
