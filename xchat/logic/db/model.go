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
	Ext     string    `db:"ext" json:"ext"`
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}

// MarshalJSON encoding this to json.
func (d *Chat) MarshalJSON() ([]byte, error) {
	type Alias Chat
	return json.Marshal(&struct {
		*Alias
		ID      string `json:"id"`
		Created int64  `json:"created"`
		Updated int64  `json:"updated"`
	}{
		Alias:   (*Alias)(d),
		ID:      EncodeChatIdentity(d.Type, d.ID),
		Created: d.Created.Unix(),
		Updated: d.Updated.Unix(),
	})
}

// UserChat is a user's conversation.
type UserChat struct {
	ID        uint64    `db:"id" json:"id"`
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	Tag       string    `json:"tag"`
	MsgID     uint64    `db:"msg_id" json:"msg_id"`
	Ext       string    `db:"ext" json:"ext"`
	Created   time.Time `json:"created"`
	Updated   time.Time `json:"updated"`
	User      string    `json:"user"`
	CurID     uint64    `db:"cur_id" json:"cur_id"`
	Joined    time.Time `json:"joined"`
	ExitMsgID uint64    `db:"exit_msg_id" json:"exit_msg_id"`
	IsExited  bool      `db:"is_exited" json:"is_exited"`
	Dnd       bool      `json:"dnd"`
}

// MarshalJSON encoding this to json.
func (d *UserChat) MarshalJSON() ([]byte, error) {
	type Alias UserChat
	return json.Marshal(&struct {
		*Alias
		ID      string `json:"id"`
		Created int64  `json:"created"`
		Updated int64  `json:"updated"`
		Joined  int64  `json:"joined"`
	}{
		Alias:   (*Alias)(d),
		ID:      EncodeChatIdentity(d.Type, d.ID),
		Created: d.Created.Unix(),
		Updated: d.Updated.Unix(),
		Joined:  d.Joined.Unix(),
	})
}

// FullMember is a chat member with full attributes.
type FullMember struct {
	User      string    `db:"user"`
	Joined    time.Time `db:"joined"`
	CurID     uint64    `db:"cur_id"`
	ExitMsgID uint64    `db:"exit_msg_id"`
	IsExited  bool      `db:"is_exited"`
	Dnd       bool      `db:"dnd"`
}

// Member is a chat member.
type Member struct {
	User   string    `db:"user" json:"user"`
	Joined time.Time `db:"joined" json:"joined"`
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
	ChatID   uint64 `db:"chat_id"`
	ChatType string `db:"chat_type"`
	Domain   string
	ID       uint64 `db:"id"`
	User     string `db:"uid"`
	Ts       time.Time
	Msg      string
}

// MarshalJSON encoding this to json.
func (d *Message) MarshalJSON() ([]byte, error) {
	type Alias Message
	return json.Marshal(&struct {
		*Alias
		ChatID string `json:"chat_id"`
		Ts     int64  `json:"ts"`
	}{
		Alias:  (*Alias)(d),
		ChatID: EncodeChatIdentity(d.ChatType, d.ChatID),
		Ts:     d.Ts.Unix(),
	})
}
