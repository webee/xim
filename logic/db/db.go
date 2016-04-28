package db

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var (
	db = sqlx.MustConnect("postgres", "postgres://xim:xim1234@localhost:5432/xim?sslmode=disable")
)

// App is a application using xim.
type App struct {
	Name string
	Desc string
	App  string
	Key  string
}

// Channel is a app's messaging channel.
type Channel struct {
	App     string
	Channel string
	Owner   string
}

// PrintApps prints app's info.
func PrintApps() {
	apps := []App{}
	if err := db.Select(&apps, `SELECT name, "desc", app, key FROM xim_app`); err != nil {
		log.Println(err)
	} else {
		for _, app := range apps {
			fmt.Printf("app: name=%s, desc=%s, app=%s, key=%s\n", app.Name, app.Desc, app.App, app.Key)
		}
	}
}
