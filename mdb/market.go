package mdb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/eveisesi/neo"
	"github.com/volatiletech/null"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type marketRepository struct {
	c *mongo.Collection
}

func NewMarketRepository(db *mongo.Database) neo.MarketRepository {
	return &marketRepository{
		db.Collection("prices"),
	}
}

func (r *marketRepository) Price(ctx context.Context, typeID uint, date string) (*neo.HistoricalRecord, error) {

	var price = new(neo.HistoricalRecord)

	err := r.c.FindOne(ctx, primitive.D{primitive.E{Key: "typeID", Value: typeID}, primitive.E{Key: "date", Value: date}}).Decode(price)
	return price, err

}

func (r *marketRepository) Prices(ctx context.Context, mods ...neo.Modifier) ([]*neo.HistoricalRecord, error) {

	filters := BuildFilters(mods...)
	findOptions := BuildFindOptions(mods...)

	var prices = make([]*neo.HistoricalRecord, 0)
	result, err := r.c.Find(ctx, filters, findOptions)
	if err != nil {
		return nil, err
	}

	err = result.All(ctx, &prices)

	return prices, err
}

func (r *marketRepository) CreatePrice(ctx context.Context, price *neo.HistoricalRecord) error {

	_, err := r.c.InsertOne(ctx, price)
	if err != nil {
		if !IsUniqueConstrainViolation(err) {
			return err
		}
	}

	return nil

}

func (r *marketRepository) BuiltPrice(ctx context.Context, id uint, date time.Time) (*neo.PriceBuilt, error) {

	price := &neo.PriceBuilt{}

	filter := primitive.D{
		primitive.E{Key: "typeID", Value: id},
		primitive.E{Key: "date", Value: date.Format("2006-01-02")},
	}

	err := r.c.FindOne(ctx, filter).Decode(price)

	return price, err

}

func (r *marketRepository) InsertBuiltPrice(ctx context.Context, price *neo.PriceBuilt) (*neo.PriceBuilt, error) {
	panic("not implemented")
}

func (r *marketRepository) HistoricalRecord(ctx context.Context, id uint, date time.Time, limit null.Int) ([]*neo.HistoricalRecord, error) {

	var records = make([]*neo.HistoricalRecord, 0)

	filter := primitive.D{
		primitive.E{
			Key:   "typeID",
			Value: id,
		},
		primitive.E{
			Key: "date",
			Value: primitive.D{
				primitive.E{
					Key:   "$lte",
					Value: date.Format("2006-01-02"),
				},
			},
		},
	}

	options := options.Find()
	if limit.Valid {
		options.SetLimit(int64(limit.Int))
	}

	options.SetSort(
		primitive.D{
			primitive.E{Key: "date", Value: -1},
		},
	)

	results, err := r.c.Find(ctx, filter, options)
	if err != nil {
		return nil, err
	}

	err = results.All(ctx, &records)

	return records, err

}

func (r *marketRepository) CreateHistoricalRecord(ctx context.Context, records []*neo.HistoricalRecord) ([]*neo.HistoricalRecord, error) {
	var values = make([]interface{}, 0)
	for _, record := range records {
		values = append(values, record)
	}

	_, err := r.c.InsertMany(ctx, values, options.InsertMany().SetOrdered(false))
	if err != nil {
		if !IsUniqueConstrainViolation(err) {
			return records, err
		}
	}

	return records, nil

}
