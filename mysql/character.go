package mysql

import (
	"context"
	"fmt"

	"github.com/eveisesi/neo"
	"github.com/eveisesi/neo/mysql/boiler"
	"github.com/jmoiron/sqlx"
	"github.com/volatiletech/sqlboiler/queries/qm"
)

type characterRepository struct {
	db *sqlx.DB
}

func NewCharacterRepository(db *sqlx.DB) killboard.CharacterRespository {
	return &characterRepository{
		db,
	}
}

func (r *characterRepository) Character(ctx context.Context, id uint64) (*killboard.Character, error) {

	var character = killboard.Character{}
	err := boiler.Characters(
		boiler.CharacterWhere.ID.EQ(id),
	).Bind(ctx, r.db, &character)

	return &character, err

}

func (r *characterRepository) CharactersByCharacterIDs(ctx context.Context, ids []uint64) ([]*killboard.Character, error) {

	var characters = make([]*killboard.Character, 0)
	err := boiler.Characters(
		qm.WhereIn(
			fmt.Sprintf(
				"%s IN ?",
				boiler.CharacterColumns.ID,
			),
			convertSliceUint64ToSliceInterface(ids)...,
		),
	).Bind(ctx, r.db, &characters)

	return characters, err
}
