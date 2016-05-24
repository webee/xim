package main

import (
	"xim/commons/db"
)

func initDB() {
	db.Init(args.dbDriverName, args.dbDatasourceName)
}
