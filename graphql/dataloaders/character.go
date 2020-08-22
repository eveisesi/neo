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
		Fetch: func(ids []uint64) ([]*neo.Character, []error) {

			errors := make([]error, len(ids))

			mods := []neo.Modifier{neo.InUint64{Column: "id", Value: ids}}

			rows, err := character.Characters(ctx, mods)
			if err != nil {
				errors = append(errors, err)
				return nil, errors
			}

			characterByCharacterID := map[uint64]*neo.Character{}
			for _, c := range rows {
				characterByCharacterID[c.ID] = c
			}

			characters := make([]*neo.Character, len(ids))
			for i, x := range ids {
				characters[i] = characterByCharacterID[x]
			}

			return characters, nil

		},
	})
}
