package mysql

import (
	"errors"

	sqlDriver "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var err error

var ErrEmptyTransaction = errors.New("sql transcation cannot be empty")

func Connect(conf *sqlDriver.Config) (db *sqlx.DB, err error) {

	db, err = sqlx.Open("nrmysql", conf.FormatDSN())
	if err != nil {
		err = errors.New("unable to create mysql connection")
		return
	}

	return
}

// func convertSliceUintToSliceInterface(n []uint) []interface{} {

// 	newSlice := make([]interface{}, len(n))
// 	for i, v := range n {
// 		newSlice[i] = v
// 	}

// 	return newSlice

// }

// func convertSliceUint64ToSliceInterface(n []uint64) []interface{} {

// 	newSlice := make([]interface{}, len(n))
// 	for i, v := range n {
// 		newSlice[i] = v
// 	}

// 	return newSlice

// }
