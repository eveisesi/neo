package resolvers

import (
	"context"

	"github.com/eveisesi/neo"
)

func (r *queryResolver) Search(ctx context.Context, term string) ([]*neo.SearchableEntity, error) {

	return r.Services.Fetch(ctx, term)

}
