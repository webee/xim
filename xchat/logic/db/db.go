package db

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	// use pg driver
	_ "github.com/lib/pq"
)

var (
	db *sqlx.DB
)

// InitDB init the db.
func InitDB(driverName, dataSourceName string) (close func()) {
	db = sqlx.MustConnect(driverName, dataSourceName)
	db.SetMaxIdleConns(1)
	db.SetMaxOpenConns(100)
	return func() {
		db.Close()
	}
}

// GetChatMembers returns chat's members.
func GetChatMembers(chatID uint64) (members []Member, err error) {
	err = db.Select(&members, `SELECT chat_id, "user", created, init_id, cur_id FROM xchat_member where chat_id=$1`, chatID)
	return
}

// AddGroupMembers add users to group.
func AddGroupMembers(chatID uint64, users []string) error {
	chat := Chat{}
	if err := db.Get(&chat, `SELECT msg_id FROM xchat_chat where id=$1 and type='group'`, chatID); err != nil {
		return err
	}

	for _, user := range users {
		db.Exec(`INSERT INTO xchat_member(chat_id, "user", created, init_id, cur_id) VALUES($1, $2, now(), $3, $4)`, chatID, user, chat.MsgID, chat.MsgID)
	}
	return nil
}

// IsChatMember judges whether user in a chat member.
func IsChatMember(chatID uint64, user string) (t bool, err error) {
	return t, db.Get(&t, `SELECT EXISTS(SELECT 1 FROM xchat_member where chat_id=$1 and "user"=$2)`, chatID, user)
}

// GetChatMessages get chat messages between sID and eID.
func GetChatMessages(chatID uint64, sID, eID uint64) (msgs []Message, err error) {
	err = db.Select(&msgs, `SELECT chat_id, msg_id, "uid", ts, msg FROM xchat_message where chat_id=$1 and msg_id > $2 and msg_id < $3 order by msg_id`, chatID, sID, eID)
	return
}

// NewMsg insert a new message.
func NewMsg(chatID uint64, user string, msg string) (message *Message, err error) {
	// 判断是否为会话成员
	t, err := IsChatMember(chatID, user)
	if err != nil {
		return
	}
	if !t {
		err = fmt.Errorf("no permission")
		return
	}

	// 插入消息
	tx, err := db.Beginx()
	if err != nil {
		return
	}

	message = &Message{
		ChatID: chatID,
		User:   user,
		Msg:    msg,
	}

	if err = tx.Get(message, `UPDATE xchat_chat SET msg_id=msg_id+1 where id=$1 RETURNING msg_id`, chatID); err != nil {
		return
	}

	if err = tx.Get(message, `INSERT INTO xchat_message(chat_id, msg_id, uid, ts, msg) values($1, $2, $3, now(), $4) RETURNING ts`, chatID, message.MsgID, user, msg); err != nil {
		return
	}
	if err = tx.Commit(); err != nil {
		if err = tx.Rollback(); err != nil {
			return
		}
	}
	return
}
