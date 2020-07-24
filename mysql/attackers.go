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

type killmailAttackerRepository struct {
	db *sqlx.DB
}

func NewKillmailAttackerRepository(db *sqlx.DB) neo.KillmailAttackerRepository {
	return &killmailAttackerRepository{
		db,
	}
}

func (r *killmailAttackerRepository) ByKillmailID(ctx context.Context, id uint64) ([]*neo.KillmailAttacker, error) {

	var attackers = make([]*neo.KillmailAttacker, 0)
	err := boiler.KillmailAttackers(
		boiler.KillmailAttackerWhere.KillmailID.EQ(id),
	).Bind(ctx, r.db, &attackers)

	return attackers, err

}

func (r *killmailAttackerRepository) ByKillmailIDs(ctx context.Context, ids []uint64) ([]*neo.KillmailAttacker, error) {

	var attackers = make([]*neo.KillmailAttacker, 0)
	err := boiler.KillmailAttackers(
		qm.WhereIn(
			fmt.Sprintf(
				"%s IN ?",
				boiler.KillmailAttackerColumns.KillmailID,
			),
			convertSliceUint64ToSliceInterface(ids)...,
		),
	).Bind(ctx, r.db, &attackers)

	return attackers, err

}

func (r *killmailAttackerRepository) Create(ctx context.Context, attacker *neo.KillmailAttacker) (*neo.KillmailAttacker, error) {

	var bAttacker = new(boiler.KillmailAttacker)
	err := copier.Copy(bAttacker, attacker)
	if err != nil {
		return nil, errors.Wrap(err, "failed to copy attacker to orm")
	}

	err = bAttacker.Insert(ctx, r.db, boil.Infer(), false)
	if err != nil {
		return nil, errors.Wrap(err, "failed to insert attacker into db")
	}

	err = copier.Copy(attacker, bAttacker)

	return attacker, errors.Wrap(err, "failed to copy orm to attacker")

}

func (r *killmailAttackerRepository) CreateWithTxn(ctx context.Context, txn neo.Transactioner, attacker *neo.KillmailAttacker) (*neo.KillmailAttacker, error) {

	var t = txn.(*transaction)
	var bAttacker = new(boiler.KillmailAttacker)
	err := copier.Copy(bAttacker, attacker)
	if err != nil {
		return nil, errors.Wrap(err, "failed to copy attacker to orm")
	}

	err = bAttacker.Insert(ctx, t, boil.Infer(), false)
	if err != nil {
		return nil, errors.Wrap(err, "failed to insert attacker into db")
	}

	err = copier.Copy(attacker, bAttacker)

	return attacker, errors.Wrap(err, "failed to copy orm to attacker")

}

func (r *killmailAttackerRepository) CreateBulk(ctx context.Context, attackers []*neo.KillmailAttacker) ([]*neo.KillmailAttacker, error) {

	for _, attacker := range attackers {
		var bAttacker = new(boiler.KillmailAttacker)
		err := copier.Copy(bAttacker, attacker)
		if err != nil {
			return nil, errors.Wrap(err, "failed to copy attacker to orm")
		}

		err = bAttacker.Insert(ctx, r.db, boil.Infer(), false)
		if err != nil {
			return nil, errors.Wrap(err, "failed to insert attacker into db")
		}

		err = copier.Copy(attacker, bAttacker)
		if err != nil {
			return nil, errors.Wrap(err, "failed to copy orm to attacker")
		}

	}

	return attackers, nil

}

func (r *killmailAttackerRepository) CreateBulkWithTxn(ctx context.Context, txn neo.Transactioner, attackers []*neo.KillmailAttacker) ([]*neo.KillmailAttacker, error) {

	for _, attacker := range attackers {
		_, err = r.CreateWithTxn(ctx, txn, attacker)
		if err != nil {
			return nil, errors.Wrap(err, "failed to insert attacker into db")
		}
	}

	return attackers, nil

}
