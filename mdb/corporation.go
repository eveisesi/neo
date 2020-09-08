package mdb

import (
	"context"
	"time"

	"github.com/eveisesi/neo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type corporationRepository struct {
	c *mongo.Collection
}

func NewCorporationRepository(d *mongo.Database) neo.CorporationRespository {
	return &corporationRepository{
		d.Collection("corporations"),
	}
}

func (r *corporationRepository) Corporation(ctx context.Context, id uint) (*neo.Corporation, error) {
	var corporation = new(neo.Corporation)

	err := r.c.FindOne(ctx, bson.D{primitive.E{Key: "id", Value: id}}).Decode(&corporation)
	return corporation, err

}

func (r *corporationRepository) Corporations(ctx context.Context, mods ...neo.Modifier) ([]*neo.Corporation, error) {

	pds := BuildFilters(mods...)
	pos := BuildFindOptions(mods...)

	var corporations = make([]*neo.Corporation, 0)
	result, err := r.c.Find(ctx, pds, pos)
	if err != nil {
		return nil, err
	}

	err = result.All(ctx, &corporations)
	return corporations, err

}

func (r *corporationRepository) CreateCorporation(ctx context.Context, corporation *neo.Corporation) error {

	corporation.CreatedAt = time.Now().Unix()
	corporation.UpdatedAt = time.Now().Unix()

	_, err := r.c.InsertOne(ctx, corporation)
	if err != nil {
		if !IsUniqueConstrainViolation(err) {
			return err
		}
	}
	return nil

}

func (r *corporationRepository) UpdateCorporation(ctx context.Context, id uint, corporation *neo.Corporation) error {

	corporation.UpdatedAt = time.Now().Unix()
	if corporation.CreatedAt == 0 {
		corporation.CreatedAt = time.Now().Unix()
	}

	update := primitive.D{primitive.E{Key: "$set", Value: corporation}}

	_, err := r.c.UpdateOne(ctx, primitive.D{{Key: "id", Value: id}}, update, nil)

	return err

}

func (r *corporationRepository) Expired(ctx context.Context) ([]*neo.Corporation, error) {
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

	return r.Corporations(ctx, mods...)
}

// https://www.mongodb.com/blog/post/quick-start-golang--mongodb--data-aggregation-pipeline
// TODO: Confirm this is working. Untested as of 2020-09-03
func (r *corporationRepository) MemberCountByAllianceID(ctx context.Context, id uint) (int, error) {

	matchStage := primitive.D{
		primitive.E{
			Key: "$match",
			Value: primitive.D{
				primitive.E{
					Key:   "allianceID",
					Value: id,
				},
			},
		},
	}
	groupStage := primitive.D{
		{
			Key: "$group",
			Value: primitive.D{
				primitive.E{
					Key:   "_id",
					Value: "null",
				},
				primitive.E{
					Key: "count",
					Value: primitive.D{
						primitive.E{
							Key:   "$sum",
							Value: "$memberCount",
						},
					},
				},
			},
		},
	}

	result, err := r.c.Aggregate(ctx, mongo.Pipeline{matchStage, groupStage})
	if err != nil {
		return 0, err
	}

	var output []primitive.M
	err = result.All(ctx, &output)
	if err != nil {
		return 0, err
	}

	return 0, nil
}
