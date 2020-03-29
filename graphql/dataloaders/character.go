package dataloaders

import (
	"context"

	"github.com/eveisesi/neo"
	"github.com/eveisesi/neo/graphql/dataloaders/generated"
	"github.com/eveisesi/neo/services/character"
)

func CharacterLoader(ctx context.Context, character character.Service) *generated.CharacterLoader {
	return generated.NewCharacterLoader(generated.CharacterLoaderConfig{
		Wait:     defaultWait,
		MaxBatch: defaultMaxBatch,
		Fetch: func(ids []uint64) ([]*killboard.Character, []error) {

			characters := make([]*killboard.Character, len(ids))
			errors := make([]error, len(ids))

			rows, err := character.CharactersByCharacterIDs(ctx, ids)
			if err != nil {
				errors = append(errors, err)
				return nil, errors
			}

			characterByCharacterID := map[uint64]*killboard.Character{}
			for _, c := range rows {
				characterByCharacterID[c.ID] = c
			}

			for i, x := range ids {
				characters[i] = characterByCharacterID[x]
			}

			return characters, nil

		},
	})
}
