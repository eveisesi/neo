package mdb

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/eveisesi/neo"
	"github.com/newrelic/go-agent/v3/integrations/nrmongo"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Connect(ctx context.Context, uri *url.URL) (*mongo.Client, error) {

	monitor := nrmongo.NewCommandMonitor(nil)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri.String()).SetMonitor(monitor))
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to mongo db")
	}

	err = client.Ping(ctx, readpref.PrimaryPreferred())
	if err != nil {
		return nil, errors.Wrap(err, "failed to ping mongo db")
	}

	return client, err

}

// Mongo Operators
const (
	equal            string = "$eq"
	greaterthan      string = "$gt"
	greaterthanequal string = "$gte"
	in               string = "$in"
	lessthan         string = "$lt"
	lessthanequal    string = "$lte"
	notequal         string = "$ne"
	notin            string = "$nin"
)

func BuildFilters(modifiers ...neo.Modifier) primitive.D {

	var mods = make(primitive.D, 0)
	for _, a := range modifiers {
		switch o := a.(type) {
		case neo.EqualTo:
			mods = append(mods, primitive.E{Key: o.Column, Value: primitive.D{primitive.E{Key: equal, Value: o.Value}}})
		case neo.NotEqualTo:
			mods = append(mods, primitive.E{Key: o.Column, Value: primitive.D{primitive.E{Key: notequal, Value: o.Value}}})
		case neo.GreaterThanEqualTo:
			mods = append(mods, primitive.E{Key: o.Column, Value: primitive.D{primitive.E{Key: greaterthanequal, Value: o.Value}}})
		case neo.GreaterThan:
			mods = append(mods, primitive.E{Key: o.Column, Value: primitive.D{primitive.E{Key: greaterthan, Value: o.Value}}})
		case neo.LessThan:
			mods = append(mods, primitive.E{Key: o.Column, Value: primitive.D{primitive.E{Key: lessthan, Value: o.Value}}})
		case neo.LessThanEqualTo:
			mods = append(mods, primitive.E{Key: o.Column, Value: primitive.D{primitive.E{Key: lessthanequal, Value: o.Value}}})
		case neo.In:
			mods = append(mods, primitive.E{Key: o.Column, Value: primitive.D{primitive.E{Key: in, Value: o.Values}}})
		case neo.NotIn:
			mods = append(mods, primitive.E{Key: o.Column, Value: primitive.D{primitive.E{Key: notin, Value: o.Values}}})
		case neo.EqualToTime:
			mods = append(mods, primitive.E{Key: o.Column, Value: primitive.D{primitive.E{Key: equal, Value: o.Value}}})
		case neo.NotEqualToTime:
			mods = append(mods, primitive.E{Key: o.Column, Value: primitive.D{primitive.E{Key: notequal, Value: o.Value}}})
		case neo.GreaterThanTime:
			mods = append(mods, primitive.E{Key: o.Column, Value: primitive.D{primitive.E{Key: greaterthan, Value: o.Value.String()}}})
		case neo.LessThanTime:
			time, _ := strconv.Unquote(fmt.Sprintf("ISODate('%s')", o.Value.Format("2006-01-02T15:04:05Z")))

			mods = append(mods, primitive.E{Key: o.Column, Value: primitive.E{Key: lessthan, Value: time}})
		case neo.ColValOr:

			arr := primitive.A{}
			for _, mod := range o.Values {
				arr = append(arr, BuildFilters(mod))
			}

			mods = append(mods, primitive.E{Key: "$or", Value: arr})

		}

	}

	return mods
}

func BuildFindOptions(modifiers ...neo.Modifier) *options.FindOptions {

	var opts = options.Find()
	for _, a := range modifiers {
		switch o := a.(type) {
		case neo.LimitModifier:
			opts.SetLimit(int64(o))
		case neo.OrderModifier:
			switch o.Sort {
			case neo.SortAsc:
				opts.SetSort(primitive.D{primitive.E{Key: o.Column, Value: 1}})
			case neo.SortDesc:
				opts.SetSort(primitive.D{primitive.E{Key: o.Column, Value: -1}})
			}

		}
	}

	return opts

}

const duplicateKeyError = 11000

func IsUniqueConstrainViolation(exception error) bool {

	var bwe mongo.BulkWriteException
	if errors.As(exception, &bwe) {

		if len(bwe.WriteErrors) == 0 {
			return false
		}
		for _, errs := range bwe.WriteErrors {
			if errs.Code == duplicateKeyError {
				return true
			}
		}
	}
	var we mongo.WriteException
	if errors.As(exception, &we) {
		if len(we.WriteErrors) == 0 {
			return false
		}
		for _, errs := range we.WriteErrors {
			if errs.Code == duplicateKeyError {
				return true
			}
		}
	}

	return false
}
