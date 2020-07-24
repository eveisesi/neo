package mysql

import (
	"context"
	"fmt"

	"github.com/eveisesi/neo"
	"github.com/eveisesi/neo/mysql/boiler"
	"github.com/jinzhu/copier"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/boil"
	"github.com/volatiletech/sqlboiler/queries/qm"
)

type killmailItemRepository struct {
	db *sqlx.DB
}

func NewKillmailItemRepository(db *sqlx.DB) neo.KillmailItemRepository {
	return &killmailItemRepository{
		db,
	}
}

func (r *killmailItemRepository) ByKillmailID(ctx context.Context, id uint64) ([]*neo.KillmailItem, error) {

	var items = make([]*neo.KillmailItem, 0)
	err := boiler.KillmailItems(
		boiler.KillmailItemWhere.KillmailID.EQ(id),
	).Bind(ctx, r.db, &items)

	return items, err

}

func (r *killmailItemRepository) ByKillmailIDs(ctx context.Context, ids []uint64) ([]*neo.KillmailItem, error) {

	var items = make([]*neo.KillmailItem, 0)
	err := boiler.KillmailItems(
		qm.WhereIn(
			fmt.Sprintf(
				"%s IN ?",
				boiler.KillmailItemColumns.KillmailID,
			),
			convertSliceUint64ToSliceInterface(ids)...,
		),
	).Bind(ctx, r.db, &items)

	return items, err

}

func (r *killmailItemRepository) Create(ctx context.Context, item *neo.KillmailItem) (*neo.KillmailItem, error) {

	var bItem = new(boiler.KillmailItem)
	err := copier.Copy(bItem, item)
	if err != nil {
		return nil, errors.Wrap(err, "failed to copy item to orm")
	}

	err = bItem.Insert(ctx, r.db, boil.Infer(), false)
	if err != nil {
		return nil, errors.Wrap(err, "failed to insert item into db")
	}

	err = copier.Copy(item, bItem)

	return item, errors.Wrap(err, "failed to copy orm to item")

}

func (r *killmailItemRepository) CreateWithTxn(ctx context.Context, txn neo.Transactioner, item *neo.KillmailItem) (*neo.KillmailItem, error) {

	var t = txn.(*transaction)
	var bItem = new(boiler.KillmailItem)
	err := copier.Copy(bItem, item)
	if err != nil {
		return nil, errors.Wrap(err, "failed to copy item to orm")
	}

	err = bItem.Insert(ctx, t, boil.Infer(), false)
	if err != nil {
		return nil, errors.Wrap(err, "failed to insert item into db")
	}

	err = copier.Copy(item, bItem)

	return item, errors.Wrap(err, "failed to copy orm to item")

}

func (r *killmailItemRepository) CreateBulk(ctx context.Context, items []*neo.KillmailItem) ([]*neo.KillmailItem, error) {

	for _, item := range items {
		var bItem = new(boiler.KillmailItem)
		err := copier.Copy(bItem, item)
		if err != nil {
			return nil, errors.Wrap(err, "failed to copy item to orm")
		}

		err = bItem.Insert(ctx, r.db, boil.Infer(), false)
		if err != nil {
			return nil, errors.Wrap(err, "failed to insert item into db")
		}

		err = copier.Copy(item, bItem)
		if err != nil {
			return nil, errors.Wrap(err, "failed to copy orm to item")
		}

	}

	return items, nil

}

func (r *killmailItemRepository) CreateBulkWithTxn(ctx context.Context, txn neo.Transactioner, items []*neo.KillmailItem) ([]*neo.KillmailItem, error) {

	var t = txn.(*transaction)
	for _, item := range items {
		var bItem = new(boiler.KillmailItem)
		err := copier.Copy(bItem, item)
		if err != nil {
			return nil, errors.Wrap(err, "failed to copy item to orm")
		}

		err = bItem.Insert(ctx, t, boil.Infer(), false)
		if err != nil {
			return nil, errors.Wrap(err, "failed to insert item into db")
		}

		err = copier.Copy(item, bItem)
		if err != nil {
			return nil, errors.Wrap(err, "failed to copy orm to item")
		}

	}

	return items, nil

}

func (r *killmailItemRepository) UpdateBulk(ctx context.Context, items []*neo.KillmailItem) error {

	for _, item := range items {
		var bItem = new(boiler.KillmailItem)
		err := copier.Copy(bItem, item)
		if err != nil {
			return errors.Wrap(err, "failed to copy item to orm")
		}

		_, err = bItem.Update(ctx, r.db, boil.Infer())
		if err != nil {
			return errors.Wrap(err, "failed to update item in db")
		}

		err = copier.Copy(item, bItem)
		if err != nil {
			return errors.Wrap(err, "failed to copy orm to item")
		}

	}

	return nil

}

func (r *killmailItemRepository) UpdateBulkWithTxn(ctx context.Context, txn neo.Transactioner, items []*neo.KillmailItem) error {

	var t = txn.(*transaction)
	for _, item := range items {
		var bItem = new(boiler.KillmailItem)
		err := copier.Copy(bItem, item)
		if err != nil {
			return errors.Wrap(err, "failed to copy item to orm")
		}

		_, err = bItem.Update(ctx, t, boil.Infer())
		if err != nil {
			return errors.Wrap(err, "failed to update item in db")
		}

		err = copier.Copy(item, bItem)
		if err != nil {
			return errors.Wrap(err, "failed to copy orm to item")
		}

	}

	return nil

}
