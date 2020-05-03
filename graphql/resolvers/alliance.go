package resolvers

import (
	"context"

	"github.com/eveisesi/neo"
)

func (r *queryResolver) AllianceByAllianceID(ctx context.Context, id int) (*neo.Alliance, error) {
	return r.Services.Alliance.Alliance(ctx, uint64(id))
}

func (r *queryResolver) KillmailsByAllianceID(ctx context.Context, allianceID int, page *int) ([]*neo.Killmail, error) {
	return r.Services.KillmailsByAllianceID(ctx, uint64(allianceID), *page)
}
