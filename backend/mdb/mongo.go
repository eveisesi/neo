package mdb

import (
	"context"
	"fmt"
	"net/url"

	"github.com/eveisesi/neo"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Connect(ctx context.Context, uri *url.URL) (*mongo.Client, error) {

	// monitor := nrmongo.NewCommandMonitor(nil)
	// .SetMonitor(monitor)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri.String()))
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to mongo db")
	}

	err = client.Ping(ctx, nil)
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
	and              string = "$and"
	or               string = "$or"
	exists           string = "$exists"
)

func BuildFilters(operators ...*neo.Operator) primitive.D {

	var ops = make(primitive.D, 0)
	for _, a := range operators {
		switch a.Operation {
		case neo.EqualOp:
			ops = append(ops, primitive.E{Key: a.Column, Value: primitive.D{primitive.E{Key: equal, Value: a.Value}}})
		case neo.NotEqualOp:
			ops = append(ops, primitive.E{Key: a.Column, Value: primitive.D{primitive.E{Key: notequal, Value: a.Value}}})
		case neo.GreaterThanOp:
			ops = append(ops, primitive.E{Key: a.Column, Value: primitive.D{primitive.E{Key: greaterthan, Value: a.Value}}})
		case neo.GreaterThanEqualToOp:
			ops = append(ops, primitive.E{Key: a.Column, Value: primitive.D{primitive.E{Key: greaterthanequal, Value: a.Value}}})
		case neo.LessThanOp:
			ops = append(ops, primitive.E{Key: a.Column, Value: primitive.D{primitive.E{Key: lessthan, Value: a.Value}}})
		case neo.LessThanEqualToOp:
			ops = append(ops, primitive.E{Key: a.Column, Value: primitive.D{primitive.E{Key: lessthanequal, Value: a.Value}}})
		case neo.ExistsOp:
			ops = append(ops, primitive.E{Key: a.Column, Value: primitive.D{primitive.E{Key: exists, Value: a.Value.(bool)}}})
		case neo.OrOp:
			switch o := a.Value.(type) {
			case []*neo.Operator:
				arr := make(primitive.A, 0)

				for _, op := range o {
					arr = append(arr, BuildFilters(op))
				}

				ops = append(ops, primitive.E{Key: or, Value: arr})
			default:
				panic(fmt.Sprintf("valid type %#T supplied, expected one of [[]*neo.Operator]", o))
			}

		case neo.AndOp:
			switch o := a.Value.(type) {
			case []*neo.Operator:
				arr := make(primitive.A, 0)
				for _, op := range o {
					arr = append(arr, BuildFilters(op))
				}

				ops = append(ops, primitive.E{Key: and, Value: arr})
			default:
				panic(fmt.Sprintf("valid type %#T supplied, expected one of [[]*neo.Operator]", o))
			}

		case neo.InOp:
			switch o := a.Value.(type) {
			case []neo.OpValue:
				arr := make(primitive.A, 0)
				for _, value := range o {
					arr = append(arr, value)
				}

				ops = append(ops, primitive.E{Key: a.Column, Value: primitive.D{primitive.E{Key: in, Value: arr}}})
			default:
				panic(fmt.Sprintf("valid type %#T supplied, expected one of [[]neo.OpValue]", o))
			}
		case neo.NotInOp:
			switch o := a.Value.(type) {
			case []neo.OpValue:
				arr := make(primitive.A, 0)
				for _, value := range o {
					arr = append(arr, value)
				}

				ops = append(ops, primitive.E{Key: a.Column, Value: primitive.D{primitive.E{Key: notin, Value: arr}}})
			default:
				panic(fmt.Sprintf("valid type %#T supplied, expected one of [[]neo.OpValue]", o))
			}
		}
	}

	return ops

}

func BuildFindOptions(ops ...*neo.Operator) *options.FindOptions {
	var opts = options.Find()
	for _, a := range ops {
		switch a.Operation {
		case neo.LimitOp:
			opts.SetLimit(a.Value.(int64))
		case neo.SkipOp:
			opts.SetSkip(a.Value.(int64))
		case neo.OrderOp:
			opts.SetSort(primitive.D{primitive.E{Key: a.Column, Value: a.Value}})
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
