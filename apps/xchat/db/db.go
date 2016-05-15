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
}

// Chat is a conversation.
type Chat struct {
	ID      uint64 `db:"id"`
	Type    string
	Channel string
	Title   string
}

// Member is the chat member.
type Member struct {
	ChatID uint64 `db:"chat_id"`
	User   string
}

// GetChatByChannel get chat by channel name.
func GetChatByChannel(channel string) (*Chat, error) {
	chat := Chat{}
	if err := db.Get(&chat, `SELECT id, type, channel, title FROM xchat_chat where channel=$1`, channel); err != nil {
		return nil, err
	}
	return &chat, nil
}

// GetChannelByChatIDAndUser get channel by chat id and user.
func GetChannelByChatIDAndUser(chatID uint64, user string) (string, error) {
	var channel string
	err := db.Get(&channel, `SELECT c.channel FROM xchat_chat c left join xchat_member m on c.id = m.chat_id where c.id=$1 and m.user=$2`, chatID, user)
	return channel, err
}
