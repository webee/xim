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
	db.SetMaxIdleConns(3)
	db.SetMaxOpenConns(100)
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
