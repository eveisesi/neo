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

type allianceRepository struct {
	db *sqlx.DB
}

func NewAllianceRepository(db *sqlx.DB) neo.AllianceRespository {
	return &allianceRepository{
		db,
	}
}

func (r *allianceRepository) Alliance(ctx context.Context, id uint64) (*neo.Alliance, error) {

	var alliance = neo.Alliance{}
	err := boiler.Alliances(
		boiler.AllianceWhere.ID.EQ(id),
	).Bind(ctx, r.db, &alliance)

	return &alliance, err

}

func (r *allianceRepository) Expired(ctx context.Context) ([]*neo.Alliance, error) {

	var alliances = make([]*neo.Alliance, 0)
	err := boiler.Alliances(
		boiler.AllianceWhere.CachedUntil.LT(time.Now()),
		qm.OrderBy(boiler.AllianceColumns.CachedUntil+" ASC"),
		qm.Limit(1000),
	).Bind(ctx, r.db, &alliances)

	return alliances, err

}

func (r *allianceRepository) CreateAlliance(ctx context.Context, alliance *neo.Alliance) (*neo.Alliance, error) {

	var bAlliance = new(boiler.Alliance)
	err := copier.Copy(bAlliance, alliance)
	if err != nil {
		return alliance, errors.Wrap(err, "unable to copy alliance to orm")
	}

	err = bAlliance.Insert(ctx, r.db, boil.Infer())
	if err != nil {
		return alliance, errors.Wrap(err, "unable to insert alliance into db")
	}

	err = copier.Copy(alliance, bAlliance)

	return alliance, errors.Wrap(err, "unable to copy orm to alliance")

}

func (r *allianceRepository) UpdateAlliance(ctx context.Context, id uint64, alliance *neo.Alliance) (*neo.Alliance, error) {

	var bAlliance = new(boiler.Alliance)
	err := copier.Copy(bAlliance, alliance)
	if err != nil {
		return alliance, errors.Wrap(err, "unable to copy alliance to orm")
	}

	bAlliance.ID = id

	_, err = bAlliance.Update(ctx, r.db, boil.Infer())
	if err != nil {
		return alliance, errors.Wrap(err, "unable to insert alliance in db")
	}

	err = copier.Copy(alliance, bAlliance)

	return alliance, errors.Wrap(err, "unable to copy orm to alliance")

}

func (r *allianceRepository) AlliancesByAllianceIDs(ctx context.Context, ids []uint64) ([]*neo.Alliance, error) {

	var alliances = make([]*neo.Alliance, 0)
	err := boiler.Alliances(
		qm.WhereIn(
			fmt.Sprintf(
				"%s IN ?",
				boiler.AllianceColumns.ID,
			),
			convertSliceUint64ToSliceInterface(ids)...,
		),
	).Bind(ctx, r.db, &alliances)

	return alliances, err
}
