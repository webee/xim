package main

import (
	"xim/apps/xchat/db"
)

func initDB() {
	db.InitDB(args.dbDriverName, args.dbDatasourceName)
}
