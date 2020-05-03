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

type killmailVictimRepository struct {
	db *sqlx.DB
}

func NewKillmailVictimRepository(db *sqlx.DB) neo.KillmailVictimRepository {
	return &killmailVictimRepository{
		db,
	}
}

func (r *killmailVictimRepository) ByKillmailID(ctx context.Context, id uint64) (*neo.KillmailVictim, error) {

	var victim = new(neo.KillmailVictim)
	err := boiler.KillmailVictims(
		boiler.KillmailVictimWhere.KillmailID.EQ(id),
	).Bind(ctx, r.db, victim)

	return victim, err

}

func (r *killmailVictimRepository) ByKillmailIDs(ctx context.Context, ids []uint64) ([]*neo.KillmailVictim, error) {

	var victims = make([]*neo.KillmailVictim, 0)
	err := boiler.KillmailVictims(
		qm.WhereIn(
			fmt.Sprintf(
				"%s IN ?",
				boiler.KillmailVictimColumns.KillmailID,
			),
			convertSliceUint64ToSliceInterface(ids)...,
		),
	).Bind(ctx, r.db, &victims)

	return victims, err

}

func (r *killmailVictimRepository) Create(ctx context.Context, victim *neo.KillmailVictim) (*neo.KillmailVictim, error) {

	var bVictim = new(boiler.KillmailVictim)
	err := copier.Copy(bVictim, victim)
	if err != nil {
		return nil, errors.Wrap(err, "failed to copy victim to orm")
	}

	err = bVictim.Insert(ctx, r.db, boil.Infer())
	if err != nil {
		return nil, errors.Wrap(err, "failed to insert victim into db")
	}

	err = copier.Copy(victim, bVictim)

	return victim, errors.Wrap(err, "failed to copy orm to victim")

}

func (r *killmailVictimRepository) CreateWithTxn(ctx context.Context, txn neo.Transactioner, victim *neo.KillmailVictim) (*neo.KillmailVictim, error) {

	var t = txn.(*transaction)
	var bVictim = new(boiler.KillmailVictim)
	err := copier.Copy(bVictim, victim)
	if err != nil {
		return nil, errors.Wrap(err, "failed to copy victim to orm")
	}

	err = bVictim.Insert(ctx, t, boil.Infer())
	if err != nil {
		return nil, errors.Wrap(err, "failed to insert victim into db")
	}

	err = copier.Copy(victim, bVictim)

	return victim, errors.Wrap(err, "failed to copy orm to victim")

}

func (r *killmailVictimRepository) Update(ctx context.Context, victim *neo.KillmailVictim) error {
	var bVictim = new(boiler.KillmailVictim)
	err := copier.Copy(bVictim, victim)
	if err != nil {
		return errors.Wrap(err, "failed to copy victim to orm")
	}

	_, err = bVictim.Update(ctx, r.db, boil.Infer())
	if err != nil {
		return errors.Wrap(err, "failed to update victim in db")
	}

	err = copier.Copy(victim, bVictim)

	return errors.Wrap(err, "failed to copy orm to victim")
}

func (r *killmailVictimRepository) UpdateWithTxn(ctx context.Context, txn neo.Transactioner, victim *neo.KillmailVictim) error {

	var t = txn.(*transaction)
	var bVictim = new(boiler.KillmailVictim)
	err := copier.Copy(bVictim, victim)
	if err != nil {
		return errors.Wrap(err, "failed to copy victim to orm")
	}

	_, err = bVictim.Update(ctx, t, boil.Infer())
	if err != nil {
		return errors.Wrap(err, "failed to update victim in db")
	}

	err = copier.Copy(victim, bVictim)

	return errors.Wrap(err, "failed to copy orm to victim")

}
