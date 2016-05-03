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

// GetChannelSubscribers get channel's subscribers.
func GetChannelSubscribers(app, channel string) []*userds.UserIdentity {
	res := make([]*userds.UserIdentity, 0, 10)
	rows, err := db.Queryx(`SELECT m.user as user FROM xim_member m left join xim_channel c on c.id = m.channel_id where c.channel=$1 and m.can_sub=true`, channel)
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
