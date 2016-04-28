package db

import (
	"fmt"
	"log"

	"database/sql"

	"github.com/jmoiron/sqlx"
	// use pg driver
	_ "github.com/lib/pq"
)

var (
	db = sqlx.MustConnect("postgres", "postgres://xim:xim1234@localhost:5432/xim?sslmode=disable")
)

// App is a application using xim.
type App struct {
	Name     string
	Desc     string
	App      string
	Password sql.NullString
}

// Channel is a app's messaging channel.
type Channel struct {
	App     string
	Channel string
	Owner   string
}

// GetApp get app by app id.
func GetApp(app string) (*App, error) {
	ximApp := App{}
	if err := db.Get(&ximApp, `SELECT name, "desc", app, password FROM xim_app where app=$1`, app); err != nil {
		return nil, err
	}
	return &ximApp, nil
}

// PrintApps prints app's info.
func PrintApps() {
	apps := []App{}
	if err := db.Select(&apps, `SELECT name, "desc", app, password FROM xim_app`); err != nil {
		log.Println(err)
	} else {
		for _, app := range apps {
			fmt.Printf("app: name=%q, desc=%q, app=%q, password=%q\n", app.Name, app.Desc, app.App, app.Password.String)
		}
	}
}
