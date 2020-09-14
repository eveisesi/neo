package mdb

import (
	"context"
	"time"

	"github.com/eveisesi/neo"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type killmailRepository struct {
	killmails  *mongo.Collection
	killhashes *mongo.Collection
}

func NewKillmailRepository(d *mongo.Database) neo.KillmailRepository {
	return &killmailRepository{
		d.Collection("killmails"),
		d.Collection("killhashes"),
	}
}

func (r *killmailRepository) Killmail(ctx context.Context, id uint) (*neo.Killmail, error) {

	var killmail = new(neo.Killmail)

	err := r.killmails.FindOne(ctx, primitive.D{primitive.E{Key: "id", Value: id}}).Decode(killmail)

	return killmail, err

}

func (r *killmailRepository) CountKillmails(ctx context.Context, operators ...*neo.Operator) (int64, error) {

	filters := BuildFilters(operators...)

	count, err := r.killmails.CountDocuments(ctx, filters)
	if err != nil {
		return 0, err
	}

	return count, err

}

func (r *killmailRepository) Killmails(ctx context.Context, operators ...*neo.Operator) ([]*neo.Killmail, error) {

	filters := BuildFilters(operators...)
	options := BuildFindOptions(operators...)

	var killmails = make([]*neo.Killmail, 0)
	result, err := r.killmails.Find(ctx, filters, options)
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

func (r *killmailRepository) Exists(ctx context.Context, id uint) (bool, error) {

	count, err := r.killmails.CountDocuments(ctx, primitive.D{primitive.E{Key: "id", Value: id}})
	if err != nil {
		return false, err
	}

	return count > 0, nil

}

func (r *killmailRepository) KillHashesByDate(ctx context.Context, date time.Time) ([]*neo.KillHash, error) {

	filters := BuildFilters(
		neo.NewEqualOperator("Date", date),
	)

	var hashes = make([]*neo.KillHash, 0)
	result, err := r.killhashes.Find(ctx, filters)
	if err != nil {
		return hashes, err
	}

	err = result.All(ctx, &hashes)

	return hashes, err
}

func (r *killmailRepository) CreateHash(ctx context.Context, hash *neo.KillHash) error {

	_, err := r.killhashes.InsertOne(ctx, hash)
	if err != nil {
		if !IsUniqueConstrainViolation(err) {
			return err
		}
	}

	return nil

}

func (r *killmailRepository) DeleteHashesByDate(ctx context.Context, date time.Time) error {

	filters := BuildFilters(
		neo.NewEqualOperator("Date", date),
	)

	_, err := r.killhashes.DeleteMany(ctx, filters)
	if err != nil {
		return err
	}

	return nil

}
