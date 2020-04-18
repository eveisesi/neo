package mysql

import (
	"github.com/eveisesi/neo"
	"github.com/jmoiron/sqlx"
)

type (
	starter struct {
		db *sqlx.DB
	}

	transaction struct {
		*sqlx.Tx
	}
)

func NewTransactioner(db *sqlx.DB) neo.Starter {
	return &starter{db}
}

func (r *starter) Begin() (neo.Transactioner, error) {

	n, err := r.db.Beginx()
	if err != nil {
		return &transaction{}, err
	}

	return &transaction{n}, nil

}
