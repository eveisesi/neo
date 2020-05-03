package resolvers

import (
	"context"

	"github.com/eveisesi/neo"
	"github.com/eveisesi/neo/graphql/service"
)

func (r *queryResolver) CharacterByCharacterID(ctx context.Context, id int) (*neo.Character, error) {
	return r.Services.Character.Character(ctx, uint64(id))
}

func (r *queryResolver) KillmailsByCharacterID(ctx context.Context, characterID int, page *int) ([]*neo.Killmail, error) {
	return r.Services.KillmailsByCharacterID(ctx, uint64(characterID), *page)
}

func (r *Resolver) Character() service.CharacterResolver {
	return &characterResolver{r}
}

type characterResolver struct{ *Resolver }

func (r *characterResolver) Corporation(ctx context.Context, obj *neo.Character) (*neo.Corporation, error) {
	return r.Dataloader(ctx).CorporationLoader.Load(obj.CorporationID)
}
