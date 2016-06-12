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
func InitDB(driverName, dataSourceName string, maxConn int) (close func()) {
	db = sqlx.MustConnect(driverName, dataSourceName)
	db.SetMaxIdleConns(1)
	db.SetMaxOpenConns(maxConn)
	return func() {
		db.Close()
	}
}

// GetChatMembers returns chat's members.
func GetChatMembers(chatID uint64) (members []Member, err error) {
	err = db.Select(&members, `SELECT chat_id, "user", joined, cur_id FROM xchat_member where chat_id=$1`, chatID)
	return
}

// AddGroupMembers add users to group.
func AddGroupMembers(chatID uint64, users []string) error {
	chat := Chat{}
	if err := db.Get(&chat, `SELECT msg_id FROM xchat_chat where id=$1 and type='group'`, chatID); err != nil {
		return err
	}

	for _, user := range users {
		db.Exec(`INSERT INTO xchat_member(chat_id, "user", created, cur_id) VALUES($1, $2, now(), $4)`, chatID, user, chat.MsgID)
	}
	return nil
}

// IsChatMember judges whether user in a chat member.
func IsChatMember(chatID uint64, user string) (t bool, err error) {
	return t, db.Get(&t, `SELECT EXISTS(SELECT 1 FROM xchat_member where chat_id=$1 and "user"=$2)`, chatID, user)
}

// GetChatMessages get chat messages between sID and eID.
func GetChatMessages(chatID uint64, lID, rID uint64, limit int, desc bool) (msgs []Message, err error) {
	// TODO: use sql generator.
	if !desc {
		if rID > 0 {
			if limit > 0 {
				err = db.Select(&msgs, `SELECT chat_id, id, uid, ts, msg FROM xchat_message WHERE chat_id=$1 and id > $2 and id < $3 ORDER BY id LIMIT $4`, chatID, lID, rID, limit)
			} else {
				err = db.Select(&msgs, `SELECT chat_id, id, uid, ts, msg FROM xchat_message WHERE chat_id=$1 and id > $2 and id < $3 ORDER BY id`, chatID, lID, rID)
			}
		} else {
			if limit > 0 {
				err = db.Select(&msgs, `SELECT chat_id, id, uid, ts, msg FROM xchat_message WHERE chat_id=$1 and id > $2 ORDER BY id LIMIT $3`, chatID, lID, limit)
			} else {
				err = db.Select(&msgs, `SELECT chat_id, id, uid, ts, msg FROM xchat_message WHERE chat_id=$1 and id > $2 ORDER BY id`, chatID, lID)
			}
		}
	} else {
		if rID > 0 {
			if limit > 0 {
				err = db.Select(&msgs, `SELECT chat_id, id, uid, ts, msg FROM xchat_message WHERE chat_id=$1 and id > $2 and id < $3 ORDER BY id DESC LIMIT $4`, chatID, lID, rID, limit)
			} else {
				err = db.Select(&msgs, `SELECT chat_id, id, uid, ts, msg FROM xchat_message WHERE chat_id=$1 and id > $2 and id < $3 ORDER BY id DESC`, chatID, lID, rID)
			}
		} else {
			if limit > 0 {
				err = db.Select(&msgs, `SELECT chat_id, id, uid, ts, msg FROM xchat_message WHERE chat_id=$1 and id > $2 ORDER BY id DESC LIMIT $3`, chatID, lID, limit)
			} else {
				err = db.Select(&msgs, `SELECT chat_id, id, uid, ts, msg FROM xchat_message WHERE chat_id=$1 and id > $2 ORDER BY id DESC`, chatID, lID)
			}
		}
	}
	return
}

// GetChat returns chat.
func GetChat(chatID uint64) (chat *Chat, err error) {
	chat = &Chat{}
	return chat, db.Get(chat, `SELECT id, type, tag, title, msg_id, created FROM xchat_chat where id=$1 and is_deleted=false`, chatID)
}

// GetUserChat returns user's chat.
func GetUserChat(user string, chatID uint64) (userChat *UserChat, err error) {
	userChat = &UserChat{}
	err = db.Get(userChat, `SELECT c.id, c.type, c.tag, c.title, c.msg_id, c.created, m.user, m.cur_id, m.joined FROM xchat_member m left join xchat_chat c on c.id = m.chat_id where m.user=$1 and c.id=$2 and c.is_deleted=false`, user, chatID)
	return
}

// GetUserChatList returns user's chat list.
func GetUserChatList(user string, onlyUnsync bool) (userChats []UserChat, err error) {
	if onlyUnsync {
		err = db.Select(&userChats, `SELECT c.id, c.type, c.tag, c.title, c.msg_id, c.created, m.user, m.cur_id, m.joined FROM xchat_chat c left join xchat_member m on c.id = m.chat_id where m.user=$1 and c.is_deleted=false and c.msg_id > m.cur_id`, user)
		return
	}
	err = db.Select(&userChats, `SELECT c.id, c.type, c.tag, c.title, c.msg_id, c.created, m.user, m.cur_id, m.joined FROM xchat_chat c left join xchat_member m on c.id = m.chat_id where m.user=$1 and c.is_deleted=false`, user)
	return
}

// SyncUserChatRecv set user's current recv msg id.
func SyncUserChatRecv(user string, chatID uint64, msgID uint64) (err error) {
	_, err = db.Exec(`UPDATE xchat_member SET cur_id=$1 WHERE "user"=$2 and chat_id=$3 and cur_id<$4`, msgID, user, chatID, msgID)
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

	if err = tx.Get(message, `UPDATE xchat_chat SET msg_id=msg_id+1 where id=$1 RETURNING msg_id as id`, chatID); err != nil {
		return
	}

	if err = tx.Get(message, `INSERT INTO xchat_message(chat_id, id, uid, ts, msg) values($1, $2, $3, now(), $4) RETURNING ts`, chatID, message.ID, user, msg); err != nil {
		return
	}
	if err = tx.Commit(); err != nil {
		if err = tx.Rollback(); err != nil {
			return
		}
	}
	return
}
