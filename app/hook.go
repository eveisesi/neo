package app

import (
	"context"

	"github.com/eveisesi/neo"
	"github.com/go-redis/redis/v7"
	"github.com/newrelic/go-agent/v3/newrelic"
)

type ctxKey int

const (
	dsCtx ctxKey = iota
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

	ctx = context.WithValue(ctx, dsCtx, ds)

	return ctx, nil

}

func (r redisHook) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	if seg, ok := ctx.Value(dsCtx).(*newrelic.DatastoreSegment); ok {
		seg.End()
	}

	return nil
}

func (r redisHook) BeforeProcessPipeline(ctx context.Context, cmd []redis.Cmder) (context.Context, error) {
	return ctx, nil
}

func (r redisHook) AfterProcessPipeline(ctx context.Context, cmd []redis.Cmder) error {
	return nil
}
