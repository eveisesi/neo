package resolvers

import (
	"context"

	"github.com/eveisesi/neo"
)

func (r *queryResolver) AllianceByAllianceID(ctx context.Context, id int) (*neo.Alliance, error) {
	return r.Services.Alliance.Alliance(ctx, uint64(id))
}
