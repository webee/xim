package db

import (
	"log"
	"xim/broker/userds"

	"github.com/jmoiron/sqlx"
	// use pg driver
	_ "github.com/lib/pq"
)

var (
	db *sqlx.DB
)

// InitDB init the db.
func InitDB(driverName, dataSourceName string) {
	db = sqlx.MustConnect(driverName, dataSourceName)
}

// App is a application using xim.
type App struct {
	Name string
	Desc string
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
	if err := db.Get(&ximApp, `SELECT name, "desc" FROM xim_app where app=$1`, app); err != nil {
		return nil, err
	}
	return &ximApp, nil
}

// CanUserPubChannel checks whether user is a publisher of channel.
func CanUserPubChannel(user userds.UserLocation, channel string) bool {
	log.Println(user, channel)
	var can bool
	if err := db.Get(&can,
		`select true from xim_app a left join xim_channel c on a.id = c.app_id left join xim_member m on c.id = m.channel_id where a.name=$1 and c.channel=$2 and m.user=$3 and can_pub=true`,
		user.App, channel, user.User); err != nil {
		log.Println(err)
		return false
	}
	return can
}
