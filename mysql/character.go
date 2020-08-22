package mysql

import (
	"context"
	"fmt"
	"time"

	"github.com/eveisesi/neo"
	"github.com/eveisesi/neo/mysql/boiler"
	"github.com/jinzhu/copier"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/boil"
)

type characterRepository struct {
	db *sqlx.DB
}

func NewCharacterRepository(db *sqlx.DB) neo.CharacterRespository {
	return &characterRepository{
		db,
	}
}

func (r *characterRepository) Character(ctx context.Context, id uint64) (*neo.Character, error) {

	var character = neo.Character{}
	err := boiler.Characters(
		boiler.CharacterWhere.ID.EQ(id),
	).Bind(ctx, r.db, &character)

	return &character, err

}

func (r *characterRepository) Characters(ctx context.Context, mods ...neo.Modifier) ([]*neo.Character, error) {

	if len(mods) == 0 {
		return nil, fmt.Errorf("Atleast one modifier must be passed in")
	}

	characters := make([]*neo.Character, 0)
	err := boiler.Characters(BuildQueryModifiers(boiler.TableNames.Characters, mods...)...).Bind(ctx, r.db, &characters)
	return characters, err

}

func (r *characterRepository) Expired(ctx context.Context) ([]*neo.Character, error) {

	mods := []neo.Modifier{
		neo.LessThanTime{Column: "CacheUntil", Value: time.Now()},
		neo.LimitModifier(1000),
		neo.OrderModifier{Column: "CacheUntil", Sort: neo.SortAsc},
	}

	return r.Characters(ctx, mods...)
}

func (r *characterRepository) CreateCharacter(ctx context.Context, character *neo.Character) (*neo.Character, error) {

	var bCharacter = new(boiler.Character)
	err := copier.Copy(bCharacter, character)
	if err != nil {
		return character, errors.Wrap(err, "unable to copy character to orm")
	}

	err = bCharacter.Insert(ctx, r.db, boil.Infer(), true)
	if err != nil {
		return character, errors.Wrap(err, "unable to insert character into db")
	}

	err = copier.Copy(character, bCharacter)

	return character, errors.Wrap(err, "unable to copy orm to character")

}

func (r *characterRepository) UpdateCharacter(ctx context.Context, id uint64, character *neo.Character) (*neo.Character, error) {

	var bCharacter = new(boiler.Character)
	err := copier.Copy(bCharacter, character)
	if err != nil {
		return character, errors.Wrap(err, "unable to copy character to orm")
	}

	bCharacter.ID = id

	_, err = bCharacter.Update(ctx, r.db, boil.Infer())
	if err != nil {
		return character, errors.Wrap(err, "unable to update character in db")
	}

	err = copier.Copy(character, bCharacter)

	return character, errors.Wrap(err, "unable to copy orm to character")

}
