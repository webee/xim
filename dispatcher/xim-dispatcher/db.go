package main

import (
	"xim/dispatcher/db"
)

func initDB() {
	db.InitDB(args.dbDriverName, args.dbDatasourceName)
}
