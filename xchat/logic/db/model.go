package db

import (
	"encoding/json"
	"time"
)

// Chat is a conversation.
type Chat struct {
	ID      uint64    `db:"id" json:"id"`
	Type    string    `json:"type"`
	Title   string    `json:"title"`
	Tag     string    `json:"tag"`
	MsgID   uint64    `db:"msg_id" json:"msg_id"`
	Created time.Time `json:"created"`
}

// MarshalJSON encoding this to json.
func (d *Chat) MarshalJSON() ([]byte, error) {
	type Alias Chat
	return json.Marshal(&struct {
		*Alias
		Created int64 `json:"created"`
	}{
		Alias:   (*Alias)(d),
		Created: d.Created.Unix(),
	})
}

// UserChat is a user's conversation.
type UserChat struct {
	ID      uint64    `db:"id" json:"id"`
	Type    string    `json:"type"`
	Title   string    `json:"title"`
	Tag     string    `json:"tag"`
	MsgID   uint64    `db:"msg_id" json:"msg_id"`
	Created time.Time `json:"created"`
	User    string    `json:"user"`
	CurID   uint64    `db:"cur_id" json:"cur_id"`
	Joined  time.Time `json:"joined"`
}

// MarshalJSON encoding this to json.
func (d *UserChat) MarshalJSON() ([]byte, error) {
	type Alias UserChat
	return json.Marshal(&struct {
		*Alias
		Created int64 `json:"created"`
		Joined  int64 `json:"joined"`
	}{
		Alias:   (*Alias)(d),
		Created: d.Created.Unix(),
		Joined:  d.Joined.Unix(),
	})
}

// Member is a chat member.
type Member struct {
	ChatID uint64 `db:"chat_id"`
	User   string
	Joined time.Time
	CurID  uint64 `db:"cur_id"`
}

// MarshalJSON encoding this to json.
func (d *Member) MarshalJSON() ([]byte, error) {
	type Alias Member
	return json.Marshal(&struct {
		*Alias
		Joined int64 `json:"joined"`
	}{
		Alias:  (*Alias)(d),
		Joined: d.Joined.Unix(),
	})
}

// Message is a chat message.
type Message struct {
	ChatID uint64 `db:"chat_id"`
	ID     uint64 `db:"id"`
	User   string `db:"uid"`
	Ts     time.Time
	Msg    string
}

// MarshalJSON encoding this to json.
func (d *Message) MarshalJSON() ([]byte, error) {
	type Alias Message
	return json.Marshal(&struct {
		*Alias
		Ts int64 `json:"ts"`
	}{
		Alias: (*Alias)(d),
		Ts:    d.Ts.Unix(),
	})
}
