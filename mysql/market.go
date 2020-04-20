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
	"github.com/volatiletech/sqlboiler/queries/qm"
)

type marketRepository struct {
	db *sqlx.DB
}

func NewMarketRepository(db *sqlx.DB) neo.MarketRepository {
	return &marketRepository{db}
}

func (r *marketRepository) Orders(ctx context.Context, id uint64) ([]*neo.Order, error) {

	var orders = make([]*neo.Order, 0)
	err := boiler.Orders(
		boiler.OrderWhere.TypeID.EQ(id),
	).Bind(ctx, r.db, &orders)

	return orders, err

}

func (r *marketRepository) OrderByTime(ctx context.Context, id uint64, minDate, maxDate time.Time) (*neo.Order, error) {

	order := new(neo.Order)
	err := boiler.Orders(
		boiler.OrderWhere.TypeID.EQ(id),
		qm.Where("date BETWEEN ? AND ?", minDate, maxDate),
		qm.OrderBy(boiler.OrderColumns.Date+" DESC"),
		qm.Limit(1),
	).Bind(context.Background(), r.db, order)

	return order, err

}

func (r *marketRepository) OrdersByIDs(ctx context.Context, ids []uint64) ([]*neo.Order, error) {

	var orders = make([]*neo.Order, 0)
	err := boiler.Orders(
		qm.WhereIn(
			fmt.Sprintf(
				"%s IN ?",
				boiler.OrderWhere.TypeID,
			),
			convertSliceUint64ToSliceInterface(ids)...,
		),
	).Bind(ctx, r.db, &orders)

	return orders, err
}

func (r *marketRepository) CreateOrdersBulk(ctx context.Context, txn neo.Transactioner, orders []*neo.Order) ([]*neo.Order, error) {

	var t = txn.(*transaction)
	for _, v := range orders {
		var order = new(boiler.Order)

		err := copier.Copy(order, v)
		if err != nil {
			txnErr := txn.Rollback()
			if txnErr != nil {
				err = errors.Wrap(err, "failed to rollback txn")
			}
			return nil, errors.Wrap(err, "failed to copy order to boiler")
		}

		err = order.Insert(ctx, t, boil.Infer())
		if err != nil {
			txnErr := txn.Rollback()
			if txnErr != nil {
				err = errors.Wrap(err, "failed to rollback txn")
			}
			return nil, errors.Wrap(err, "failed to insert order")
		}

		err = copier.Copy(v, order)
		if err != nil {
			txnErr := txn.Rollback()
			if txnErr != nil {
				err = errors.Wrap(err, "failed to rollback txn")
			}
			return nil, errors.Wrap(err, "failed to copy inserted record back to neo")
		}

	}

	return orders, nil

}
