package main

import (
	"xim/commons/db"
	"xim/commons/msgdb"
)

func initDB() {
	db.Init(args.dbDriverName, args.dbDatasourceName)
	msgdb.Init(args.mangoURL)
}
