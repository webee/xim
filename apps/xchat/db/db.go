package db

import (
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

// GetChatByChannel get chat by channel name.
func GetChatByChannel(channel string) (*Chat, error) {
	chat := Chat{}
	if err := db.Get(&chat, `SELECT id, type, channel FROM xchat_chat where channel=$1`, channel); err != nil {
		return nil, err
	}
	return &chat, nil
}

// GetChannelByChatIDAndUser get channel by chat id and user.
func GetChannelByChatIDAndUser(chatID uint64, user string) (string, error) {
	var channel string
	err := db.Get(&channel, `SELECT c.channel FROM xchat_chat c left join xchat_member m on c.id = m.chat_id left join xchat_user u on u.id = m.user_id where c.id=$1 and u.user=$2`, chatID, user)
	return channel, err
}

// GetMemberInfoByChatIDAndUser get member info by chat id and user.
func GetMemberInfoByChatIDAndUser(chatID uint64, user string) (*MemberInfo, error) {
	info := MemberInfo{}
	err := db.Get(&info, `SELECT c.channel, m.init_id FROM xchat_chat c left join xchat_member m on c.id = m.chat_id left join xchat_user u on u.id = m.user_id where c.id=$1 and u.user=$2`, chatID, user)
	return &info, err
}
