package db

import (
	"github.com/jmoiron/sqlx"
	// use pg driver
	_ "github.com/lib/pq"
)

var (
	db *sqlx.DB
)

// InitDB init the db.
func InitDB(driverName, dataSourceName string, maxConn int) (close func()) {
	db = sqlx.MustConnect(driverName, dataSourceName)
	db.SetMaxIdleConns(1)
	db.SetMaxOpenConns(maxConn)
	return func() {
		db.Close()
	}
}

