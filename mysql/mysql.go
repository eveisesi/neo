package mysql

import (
	"database/sql"
	"errors"

	sqlDriver "github.com/go-sql-driver/mysql"
)

func Connect(conf *sqlDriver.Config) (db *sql.DB, err error) {

	db, err = sql.Open("mysql", conf.FormatDSN())
	if err != nil {
		err = errors.New("unable to create mysql connection")
		return
	}

	return
}
