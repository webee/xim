package db

import (
	"fmt"
	"time"
)

// Timestamp is timestamp.
type Timestamp time.Time

// Unix returns unix timestamp.
func (t Timestamp) Unix() int64 {
	return time.Time(t).Unix()
}

// MarshalJSON encode Timestamp to byte array.
func (t Timestamp) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%d", time.Time(t).Unix())), nil
}

// Chat is a conversation.
type Chat struct {
	ID      uint64 `db:"id"`
	Type    string
	Title   string
	Tag     string
	MsgID   uint64 `db:"msg_id"`
	Created time.Time
}

// UserChat is a user's conversation.
type UserChat struct {
	ID      uint64    `db:"id" json:"id"`
	Type    string    `json:"type"`
	Title   string    `json:"title"`
	Tag     string    `json:"tag"`
	MsgID   uint64    `db:"msg_id" json:"msg_id"`
	Created Timestamp `json:"created"`
	User    string    `json:"user"`
	CurID   uint64    `db:"cur_id" json:"cur_id"`
	Joined  Timestamp `json:"joined"`
}

// Member is a chat member.
type Member struct {
	ChatID  uint64 `db:"chat_id"`
	User    string
	Created time.Time
	CurID   uint64 `db:"cur_id"`
}

// Message is a chat message.
type Message struct {
	ChatID uint64 `db:"chat_id"`
	ID     uint64 `db:"id"`
	User   string `db:"uid"`
	Ts     time.Time
	Msg    string
}
