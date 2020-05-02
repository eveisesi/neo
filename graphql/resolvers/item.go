package resolvers

import (
	"context"

	"github.com/eveisesi/neo"
	"github.com/eveisesi/neo/graphql/service"
)

func (r *Resolver) KillmailItem() service.KillmailItemResolver {
	return &killmailItemResolver{r}
}

type killmailItemResolver struct{ *Resolver }

func (r *killmailItemResolver) Type(ctx context.Context, obj *neo.KillmailItem) (*neo.Type, error) {
	return r.Dataloader(ctx).TypeLoader.Load(obj.ItemTypeID)
}

func (r *killmailItemResolver) Typeflag(ctx context.Context, obj *neo.KillmailItem) (*neo.TypeFlag, error) {
	if obj.Flag == 0 {
		return nil, nil
	}
	return r.Dataloader(ctx).TypeFlagLoader.Load(obj.Flag)
}
