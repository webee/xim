package main

import (
	"xim/logic/db"
)

func initDB() {
	db.InitDB(args.dbDriverName, args.dbDatasourceName)
}
