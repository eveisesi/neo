package mdb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/eveisesi/neo"
	"go.mongodb.org/mongo-driver/mongo"
)

type characterRepository struct {
	c *mongo.Collection
}

func NewCharacterRepository(d *mongo.Database) neo.CharacterRespository {
	return &characterRepository{
		d.Collection("characters"),
	}
}

func (r *characterRepository) Character(ctx context.Context, id uint64) (*neo.Character, error) {

	character := neo.Character{}

	err := r.c.FindOne(ctx, primitive.D{primitive.E{Key: "id", Value: id}}).Decode(&character)

	return &character, err

}

func (r *characterRepository) Characters(ctx context.Context, mods ...neo.Modifier) ([]*neo.Character, error) {

	pds := BuildFilters(mods...)
	pos := BuildFindOptions(mods...)

	var characters = make([]*neo.Character, 0)
	result, err := r.c.Find(ctx, pds, pos)
	if err != nil {
		return nil, err
	}

	err = result.All(ctx, &characters)

	return characters, err
}

func (r *characterRepository) CreateCharacter(ctx context.Context, character *neo.Character) error {

	now := time.Now().Unix()
	character.CreatedAt = now
	character.UpdatedAt = now

	_, err := r.c.InsertOne(ctx, character)

	return err

}

func (r *characterRepository) UpdateCharacter(ctx context.Context, id uint64, character *neo.Character) error {

	character.UpdatedAt = time.Now().Unix()
	if character.CreatedAt == 0 {
		character.CreatedAt = time.Now().Unix()
	}

	update := primitive.D{primitive.E{Key: "$set", Value: character}}

	_, err := r.c.UpdateOne(ctx, primitive.D{{Key: "id", Value: id}}, update, nil)

	return err

}

func (r *characterRepository) DeleteCharacter(ctx context.Context, id uint64) error {
	panic("implement me")
}

func (r *characterRepository) Expired(ctx context.Context) ([]*neo.Character, error) {

	mods := []neo.Modifier{
		neo.LessThan{Column: "cachedUntil", Value: time.Now().Unix()},
		neo.OrMod{
			Values: []neo.Modifier{
				neo.NotExists{Column: "updateError"},
				neo.LessThan{Column: "updateError", Value: 3},
			},
		},
		neo.LimitModifier(1000),
		neo.OrderModifier{Column: "cachedUntil", Sort: neo.SortAsc},
	}

	return r.Characters(ctx, mods...)
}
