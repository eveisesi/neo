package mysql

import (
	"errors"

	sqlDriver "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

func Connect(conf *sqlDriver.Config) (db *sqlx.DB, err error) {

	db, err = sqlx.Open("mysql", conf.FormatDSN())
	if err != nil {
		err = errors.New("unable to create mysql connection")
		return
	}

	return
}
