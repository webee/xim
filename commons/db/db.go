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

// Init init the db.
func Init(driverName, dataSourceName string) {
	db = sqlx.MustConnect(driverName, dataSourceName)
	db.SetMaxIdleConns(3)
	db.SetMaxOpenConns(100)
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
		`select true from xim_channel c left join xim_channel_publishers p on c.id = p.channel_id left join xim_appuser u on u.id = p.appuser_id where c.channel=$1 and u.user=$2`,
		channel, user.User); err != nil {
		log.Println(err)
		return false
	}
	return can
}

// GetChannelSubscribers get channel's subscribers.
func GetChannelSubscribers(app, channel string) []*userds.UserIdentity {
	res := make([]*userds.UserIdentity, 0, 10)
	rows, err := db.Queryx(`SELECT u.user as user FROM xim_channel c left join xim_channel_subscribers s on c.id = s.channel_id left join xim_appuser u on u.id = s.appuser_id where c.channel=$1`, channel)
	if err != nil {
		log.Println(err)
		return res
	}

	for rows.Next() {
		uid := userds.UserIdentity{
			App: app,
		}
		if err = rows.StructScan(&uid); err != nil {
			log.Println(err)
			continue
		}
		res = append(res, &uid)
	}
	return res
}

// AppChannelExists check if app channel exists.
func AppChannelExists(app, channel string) bool {
	var exists bool
	if err := db.Get(&exists, `SELECT true FROM xim_app a left join xim_channel c on a.id = c.app_id where a.name=$1 and c.channel=$2`, app, channel); err != nil {
		log.Println(err)
		return false
	}
	return exists
}
