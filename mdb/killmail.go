package mdb

import (
	"context"

	"github.com/davecgh/go-spew/spew"
	"github.com/eveisesi/neo"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type killmailRepository struct {
	killmails *mongo.Collection
}

func NewKillmailRepository(d *mongo.Database) neo.KillmailRepository {
	return &killmailRepository{
		d.Collection("killmails"),
	}
}

func (r *killmailRepository) Killmail(ctx context.Context, id uint) (*neo.Killmail, error) {

	var killmail = new(neo.Killmail)

	err := r.killmails.FindOne(ctx, primitive.D{primitive.E{Key: "id", Value: id}}).Decode(killmail)

	return killmail, err

}

func (r *killmailRepository) Killmails(ctx context.Context, mods ...neo.Modifier) ([]*neo.Killmail, error) {

	filters := BuildFilters(mods...)
	opts := BuildFindOptions(mods...)

	spew.Dump(filters, opts)

	return nil, nil

	var killmails = make([]*neo.Killmail, 0)
	result, err := r.killmails.Find(ctx, filters, opts)
	if err != nil {
		return killmails, err
	}

	err = result.All(ctx, &killmails)

	return killmails, err

}

func (r *killmailRepository) CreateKillmail(ctx context.Context, killmail *neo.Killmail) error {

	_, err := r.killmails.InsertOne(ctx, killmail)
	if err != nil {
		if !IsUniqueConstrainViolation(err) {
			return err
		}
	}

	return nil

}

func (r *killmailRepository) Recent(ctx context.Context, limit, offset int) ([]*neo.Killmail, error) {
	return nil, nil
}

func (r *killmailRepository) Exists(ctx context.Context, id uint) (bool, error) {

	count, err := r.killmails.CountDocuments(ctx, primitive.D{primitive.E{Key: "id", Value: id}})
	if err != nil {
		return false, err
	}

	return count > 0, nil

}
