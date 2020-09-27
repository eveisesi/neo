package resolvers

import (
	"context"

	"github.com/eveisesi/neo"
	"github.com/eveisesi/neo/graphql/service"
)

func (r *queryResolver) AllianceByAllianceID(ctx context.Context, id int) (*neo.Alliance, error) {
	return r.Services.Alliance.Alliance(ctx, uint(id))
}

func (r *Resolver) Alliance() service.AllianceResolver {
	return &allianceResolver{r}
}

type allianceResolver struct{ *Resolver }

func (r *allianceResolver) MemberCount(ctx context.Context, obj *neo.Alliance) (int, error) {
	return r.Services.MemberCountByAllianceID(ctx, obj.ID)
}
