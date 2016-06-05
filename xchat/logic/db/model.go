package db

import (
	"time"
)

// Chat is a conversation.
type Chat struct {
	ID        uint64 `db:"id"`
	Type      string
	Title     string
	Tag       string
	MsgID     string `db:"msg_id"`
	IsDeleted bool   `db:"is_deleted"`
	Created   time.Time
}

// Member is a chat member.
type Member struct {
	ChatID  uint64 `db:"chat_id"`
	User    string
	Created time.Time
	InitID  uint64 `db:"init_id"`
	CurID   uint64 `db:"cur_id"`
}

// Message is a chat message.
type Message struct {
	ChatID uint64 `db:"chat_id"`
	MsgID  uint64 `db:"msg_id"`
	User   string `db:"uid"`
	Ts     time.Time
	Msg    string
}
