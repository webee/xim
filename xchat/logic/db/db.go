package db

import (
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	// use pg driver
	_ "github.com/lib/pq"
)

const (
	// UserChatSQLPrefix sql prefix for querying UserChat.
	UserChatSQLPrefix = "SELECT c.id, c.type, c.tag, c.title, c.app_id, c.msg_id, c.ext, c.created, c.updated, c.members_updated, c.last_msg_ts, m.user, m.cur_id, m.joined, m.exit_msg_id, m.is_exited, m.dnd, m.label, m.updated AS user_updated, m.join_msg_id FROM xchat_chat c left join xchat_member m on c.id = m.chat_id"
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

// GetApp returns app
func GetApp(appID string) (*App, error) {
	app := &App{}
	return app, db.Get(app, `SELECT app_id, event_notify_url FROM xchat_app where app_id=$1`, appID)
}

// GetFullChatMembers returns chat's members with full attributes.
func GetFullChatMembers(chatID uint64) (members []FullMember, err error) {
	err = db.Select(&members, `SELECT "user", joined, cur_id, exit_msg_id, is_exited, dnd, label, updated FROM xchat_member where chat_id=$1 and is_exited=false`, chatID)
	return
}

// GetChatMembers returns chat's members.
func GetChatMembers(chatID uint64) (members []Member, err error) {
	err = db.Select(&members, `SELECT "user", joined FROM xchat_member where chat_id=$1 and is_exited=false`, chatID)
	return
}

// AddChatMembers add users to chat.
func AddChatMembers(chatID uint64, users []string, limit int) (err error) {
	return Transaction(db, func(tx *sqlx.Tx) error {
		if limit > 0 {
			var count int
			if err := tx.Get(&count, `SELECT count(*) FROM xchat_member WHERE chat_id=$1`, chatID); err != nil {
				return err
			}
			if count > limit {
				return errors.New("too many members")
			}
		}

		chat := Chat{}
		if err := tx.Get(&chat, `SELECT msg_id FROM xchat_chat where id=$1`, chatID); err != nil {
			return err
		}

		for _, user := range users {
			if _, err := tx.Exec(`INSERT INTO xchat_member(chat_id, "user", joined, cur_id, join_msg_id) VALUES($1, $2, now(), $3, $4)`, chatID, user, chat.MsgID, chat.MsgID); err != nil {
				return err
			}
		}
		if _, err := tx.Exec(`UPDATE xchat_chat SET updated=now(),members_updated=now() WHERE id=$1`, chatID); err != nil {
			return err
		}
		return nil
	})
}

// RemoveChatMembers removes users from chat.
func RemoveChatMembers(chatID uint64, users []string) (err error) {
	return Transaction(db, func(tx *sqlx.Tx) error {
		for _, user := range users {
			if _, err = tx.Exec(`DELETE FROM xchat_member WHERE chat_id=$1 and "user"=$2`, chatID, user); err != nil {
				return err
			}
		}
		if _, err := tx.Exec(`UPDATE xchat_chat SET updated=now(),members_updated=now() WHERE id=$1`, chatID); err != nil {
			return err
		}
		return nil
	})
}

// IsRoomChat judges whether chat is own to room.
func IsRoomChat(roomID, chatID uint64) (t bool, err error) {
	return t, db.Get(&t, `SELECT EXISTS(SELECT 1 FROM xchat_chat where room_id=$1 and id=$2)`, roomID, chatID)
}

// RoomExists judges whether the room exists.
func RoomExists(roomID uint64) (t bool, err error) {
	return t, db.Get(&t, `SELECT EXISTS(SELECT 1 FROM xchat_room where id=$1)`, roomID)
}

// IsHaveUserChat check if user1 and user2 have user chat.
func IsHaveUserChat(user1, user2 string) (t bool, err error) {
	// FIXME: use type&key to judge.
	return t, db.Get(&t, `SELECT EXISTS(select 1 from xchat_chat c where c.type='user' and exists (select 1 from xchat_member m where m.chat_id=c.id and m."user"=$1) and exists (select 1 from xchat_member m where m.chat_id=c.id and m."user"=$2))`, user1, user2)
}

// IsChatMember judges whether user in a chat member.
func IsChatMember(chatID uint64, user string) (t bool, err error) {
	return t, db.Get(&t, `SELECT EXISTS(SELECT 1 FROM xchat_member where chat_id=$1 and "user"=$2)`, chatID, user)
}

// GetChatMessages get chat messages between sID and eID.
func GetChatMessages(chatID uint64, chatType string, lID, rID uint64, limit int, desc bool) (msgs []Message, err error) {
	// TODO: use sql generator.
	if !desc {
		if rID > 0 {
			if limit > 0 {
				err = db.Select(&msgs, `SELECT chat_id, chat_type, id, uid, ts, msg, domain FROM xchat_message WHERE chat_id=$1 and chat_type=$2 and id > $3 and id < $4 ORDER BY id LIMIT $5`, chatID, chatType, lID, rID, limit)
			} else {
				err = db.Select(&msgs, `SELECT chat_id, chat_type, id, uid, ts, msg, domain FROM xchat_message WHERE chat_id=$1 and chat_type=$2 and id > $3 and id < $4 ORDER BY id`, chatID, chatType, lID, rID)
			}
		} else {
			if limit > 0 {
				err = db.Select(&msgs, `SELECT chat_id, chat_type, id, uid, ts, msg, domain FROM xchat_message WHERE chat_id=$1 and chat_type=$2 and id > $3 ORDER BY id LIMIT $4`, chatID, chatType, lID, limit)
			} else {
				err = db.Select(&msgs, `SELECT chat_id, chat_type, id, uid, ts, msg, domain FROM xchat_message WHERE chat_id=$1 and chat_type=$2 and id > $3 ORDER BY id`, chatID, chatType, lID)
			}
		}
	} else {
		if rID > 0 {
			if limit > 0 {
				err = db.Select(&msgs, `SELECT chat_id, chat_type, id, uid, ts, msg, domain FROM xchat_message WHERE chat_id=$1 and chat_type=$2 and id > $3 and id < $4 ORDER BY id DESC LIMIT $5`, chatID, chatType, lID, rID, limit)
			} else {
				err = db.Select(&msgs, `SELECT chat_id, chat_type, id, uid, ts, msg, domain FROM xchat_message WHERE chat_id=$1 and chat_type=$2 and id > $3 and id < $4 ORDER BY id DESC`, chatID, chatType, lID, rID)
			}
		} else {
			if limit > 0 {
				err = db.Select(&msgs, `SELECT chat_id, chat_type, id, uid, ts, msg, domain FROM xchat_message WHERE chat_id=$1 and chat_type=$2 and id > $3 ORDER BY id DESC LIMIT $4`, chatID, chatType, lID, limit)
			} else {
				err = db.Select(&msgs, `SELECT chat_id, chat_type, id, uid, ts, msg, domain FROM xchat_message WHERE chat_id=$1 and chat_type=$2 and id > $3 ORDER BY id DESC`, chatID, chatType, lID)
			}
		}
	}
	return
}

// GetChatMessagesByIDs get chat messages by ids.
func GetChatMessagesByIDs(chatID uint64, chatType string, msgIDs []uint64) (msgs []Message, err error) {
	if len(msgIDs) == 0 {
		return
	}

	query, args, err := sqlx.In(`SELECT chat_id, chat_type, id, uid, ts, msg, domain FROM xchat_message WHERE chat_id=? and chat_type=? and id IN (?)`, chatID, chatType, msgIDs)
	if err != nil {
		return nil, err
	}

	query = db.Rebind(query)

	err = db.Select(&msgs, query, args...)
	return
}

// GetSelfChat get user's self chat.
func GetSelfChat(user string) (*Chat, error) {
	chat := &Chat{}
	return chat, db.Get(chat, `SELECT id, type, owner, tag, title, msg_id, ext, created, updated, members_updated FROM xchat_chat where type=$1 AND key=$2`, "self", user)
}

// GetOrCreateSelfChat get or craete self chat.
func GetOrCreateSelfChat(user string) (chat *Chat, err error) {
	chat, err = GetSelfChat(user)
	if err == nil {
		return
	}

	// new self chat.
	err = Transaction(db, func(tx *sqlx.Tx) (err error) {
		var chatID uint64
		if err = tx.Get(&chatID, `INSERT INTO xchat_chat("type", key, owner, title, tag, msg_id, last_msg_ts, ext, is_deleted, created, updated, members_updated) VALUES('self', $1, $2, '', 'user', 0, '1970-01-01'::date, '', false, now(), now(), now()) RETURNING id`, user, user); err != nil {
			return err
		}
		if _, err := tx.Exec(`INSERT INTO xchat_member(chat_id, "user", joined, cur_id) VALUES($1, $2, now(), 0)`, chatID, user); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return GetSelfChat(user)
}

// GetChat returns chat.
func GetChat(chatID uint64) (chat *Chat, err error) {
	chat = &Chat{}
	return chat, db.Get(chat, `SELECT id, type, owner, tag, title, msg_id, ext, created, updated, members_updated FROM xchat_chat where id=$1 and is_deleted=false`, chatID)
}

// GetChatWithType returns chat.
func GetChatWithType(chatID uint64, chatType string) (chat *Chat, err error) {
	chat = &Chat{}
	return chat, db.Get(chat, `SELECT id, type, owner, tag, app_id, title, msg_id, ext, created, updated, members_updated FROM xchat_chat where id=$1 and type=$2 and is_deleted=false`, chatID, chatType)
}

// GetUserChat returns user's chat.
func GetUserChat(user string, chatID uint64) (userChat *UserChat, err error) {
	userChat = &UserChat{}
	err = db.Get(userChat, fmt.Sprintf("%s %s", UserChatSQLPrefix, `where m.user=$1 and c.id=$2 and c.is_deleted=false`), user, chatID)
	return
}

// GetUserChatWithType returns user's chat.
func GetUserChatWithType(user string, chatID uint64, chatType string) (userChat *UserChat, err error) {
	userChat = &UserChat{}
	err = db.Get(userChat, fmt.Sprintf("%s %s", UserChatSQLPrefix, `where m.user=$1 and c.id=$2 and c.type=$3 and c.is_deleted=false`), user, chatID, chatType)
	return
}

// GetUserChatList returns user's chat list.
func GetUserChatList(user string, onlyUnsync bool, lastMsgTs int64) (userChats []UserChat, err error) {
	if lastMsgTs <= 0 {
		if onlyUnsync {
			err = db.Select(&userChats, fmt.Sprintf("%s %s", UserChatSQLPrefix, `where m.user=$1 and c.is_deleted=false and c.msg_id > m.cur_id`), user)
		} else {
			err = db.Select(&userChats, fmt.Sprintf("%s %s", UserChatSQLPrefix, `where m.user=$1 and c.is_deleted=false`), user)
		}
	} else {
		if onlyUnsync {
			err = db.Select(&userChats, fmt.Sprintf("%s %s", UserChatSQLPrefix, `where m.user=$1 and c.is_deleted=false and c.msg_id > m.cur_id and c.last_msg_ts >= $2`), user, time.Unix(lastMsgTs+1, 0))
		} else {
			err = db.Select(&userChats, fmt.Sprintf("%s %s", UserChatSQLPrefix, `where m.user=$1 and c.is_deleted=false and c.last_msg_ts >= $2`), user, time.Unix(lastMsgTs+1, 0))
		}
	}
	return
}

// GetOrCreateNewRoomChats gets room or craete new chats.
func GetOrCreateNewRoomChats(roomID uint64, chatIDs []uint64) (roomChats []RoomChat, err error) {
	if len(chatIDs) == 0 {
		chatIDs = append(chatIDs, 0)
	}

	query, args, err := sqlx.In("SELECT rc.area, rc.chat_id FROM xchat_roomchat rc left join xchat_chat c on rc.chat_id=c.id WHERE rc.room_id=? and c.is_deleted=false and rc.chat_id NOT IN (?) ORDER BY rc.chat_id", roomID, chatIDs)
	if err != nil {
		return nil, err
	}

	query = db.Rebind(query)

	err = db.Select(&roomChats, query, args...)
	if err != nil {
		return nil, err
	}

	if len(roomChats) > 0 {
		return
	}

	// new chat.
	var chatID uint64
	var area uint32
	err = Transaction(db, func(tx *sqlx.Tx) (err error) {
		var t bool
		if err = db.Get(&t, `SELECT EXISTS(SELECT 1 FROM xchat_room where id=$1)`, roomID); err != nil {
			return err
		}
		if !t {
			return errors.New("room not exists")
		}

		if err = tx.Get(&chatID, `INSERT INTO xchat_chat("type", title, tag, msg_id, last_msg_ts, ext, is_deleted, created, updated, members_updated) VALUES('room', '$1', '_room', 0, '1970-01-01'::date, '', false, now(), now(), now()) RETURNING id`, roomID); err != nil {
			return err
		}
		if err = tx.Get(&area, `SELECT count(area) FROM xchat_roomchat WHERE room_id=$1`, roomID); err != nil {
			return err
		}
		if _, err = tx.Exec(`INSERT INTO xchat_roomchat(room_id, area, chat_id) VALUES($1, $2, $3)`, roomID, area, chatID); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	roomChats = append(roomChats, RoomChat{area, chatID})
	return
}

// SetUserChatTitle set chat's title.
func SetUserChatTitle(user string, chatID uint64, chatType string, title string) (err error) {
	_, err = db.Exec(`UPDATE xchat_chat SET title=$1, updated=now() WHERE id=$2 AND type=$3 AND EXISTS(SELECT 1 FROM xchat_member WHERE chat_id=$4 AND "user"=$5)`, title, chatID, chatType, chatID, user)
	return
}

// SetUserChat set user's chat attribute.
func SetUserChat(user string, chatID uint64, key string, value interface{}) (updated time.Time, err error) {
	s := fmt.Sprintf(`UPDATE xchat_member SET %s=$1, updated=now() WHERE "user"=$2 and chat_id=$3 RETURNING updated`, key)
	err = db.Get(&updated, s, value, user, chatID)
	return
}

// SyncUserChatRecv set user's current recv msg id.
func SyncUserChatRecv(user string, chatID uint64, msgID uint64) (err error) {
	_, err = db.Exec(`UPDATE xchat_member SET cur_id=$1 WHERE "user"=$2 and chat_id=$3 and cur_id<$4 and exists (select 1 from xchat_chat where id=$5 and msg_id >= $6)`, msgID, user, chatID, msgID, chatID, msgID)
	return
}

// NewMsg insert a new message.
func NewMsg(chatID uint64, chatType, domain string, user string, msg string) (message *Message, err error) {
	err = Transaction(db, func(tx *sqlx.Tx) error {
		message = &Message{
			ChatID:   chatID,
			ChatType: chatType,
			Domain:   domain,
			User:     user,
			Msg:      msg,
		}

		if err = tx.Get(message, `UPDATE xchat_chat SET msg_id=msg_id+1, last_msg_ts=now(), updated=now() where id=$1 and type=$2 RETURNING msg_id as id, last_msg_ts as ts`, chatID, chatType); err != nil {
			return err
		}
		if _, err = tx.Exec(`INSERT INTO xchat_message(chat_id, chat_type, id, uid, ts, msg, domain) values($1, $2, $3, $4, $5, $6, $7)`, chatID, chatType, message.ID, user, message.Ts, msg, domain); err != nil {
			return err
		}
		return nil
	})
	return
}

// Transaction is a transaction wrapper.
func Transaction(db *sqlx.DB, txFunc func(*sqlx.Tx) error) (err error) {
	tx, err := db.Beginx()
	if err != nil {
		return
	}
	defer func() {
		if p := recover(); p != nil {
			switch p := p.(type) {
			case error:
				err = p
			default:
				err = fmt.Errorf("%s", p)
			}
		}
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()
	return txFunc(tx)
}
