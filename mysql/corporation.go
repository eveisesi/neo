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

type corporationRepository struct {
	db *sqlx.DB
}

func NewCorporationRepository(db *sqlx.DB) neo.CorporationRespository {
	return &corporationRepository{
		db,
	}
}

func (r *corporationRepository) Corporation(ctx context.Context, id uint64) (*neo.Corporation, error) {

	var corporation = neo.Corporation{}
	err := boiler.Corporations(
		boiler.CorporationWhere.ID.EQ(id),
	).Bind(ctx, r.db, &corporation)

	return &corporation, err

}

func (r *corporationRepository) Expired(ctx context.Context) ([]*neo.Corporation, error) {

	var corporations = make([]*neo.Corporation, 0)
	err := boiler.Corporations(
		boiler.CorporationWhere.CachedUntil.LT(time.Now()),
		qm.OrderBy(boiler.CorporationColumns.CachedUntil+" ASC"),
		qm.Limit(1000),
	).Bind(ctx, r.db, &corporations)

	return corporations, err

}

func (r *corporationRepository) CreateCorporation(ctx context.Context, corporation *neo.Corporation) (*neo.Corporation, error) {

	var bCorporation = new(boiler.Corporation)
	err := copier.Copy(bCorporation, corporation)
	if err != nil {
		return corporation, errors.Wrap(err, "unable to copy corporation to orm")
	}

	err = bCorporation.Insert(ctx, r.db, boil.Infer())
	if err != nil {
		return corporation, errors.Wrap(err, "unable to insert corporation into db")
	}

	err = copier.Copy(corporation, bCorporation)

	return corporation, errors.Wrap(err, "unable to copy orm to corporation")

}

func (r *corporationRepository) UpdateCorporation(ctx context.Context, id uint64, corporation *neo.Corporation) (*neo.Corporation, error) {

	var bCorporation = new(boiler.Corporation)
	err := copier.Copy(bCorporation, corporation)
	if err != nil {
		return corporation, errors.Wrap(err, "unable to copy corporation to orm")
	}

	bCorporation.ID = id

	_, err = bCorporation.Update(ctx, r.db, boil.Infer())
	if err != nil {
		return corporation, errors.Wrap(err, "unable to insert corporation in db")
	}

	err = copier.Copy(corporation, bCorporation)

	return corporation, errors.Wrap(err, "unable to copy orm to corporation")

}

func (r *corporationRepository) CorporationsByCorporationIDs(ctx context.Context, ids []uint64) ([]*neo.Corporation, error) {

	var corporations = make([]*neo.Corporation, 0)
	err := boiler.Corporations(
		qm.WhereIn(
			fmt.Sprintf(
				"%s IN ?",
				boiler.CorporationColumns.ID,
			),
			convertSliceUint64ToSliceInterface(ids)...,
		),
	).Bind(ctx, r.db, &corporations)

	return corporations, err
}
