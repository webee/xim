package db

import (
	"database/sql"
	"encoding/json"
	"time"
)

// Chat is a conversation.
type Chat struct {
	ID             uint64         `db:"id" json:"id"`
	Type           string         `json:"type"`
	Owner          sql.NullString `json:"-"`
	Title          string         `json:"title"`
	Tag            string         `json:"tag"`
	MsgID          uint64         `db:"msg_id" json:"msg_id"`
	Ext            string         `db:"ext" json:"ext"`
	Created        time.Time      `json:"created"`
	Updated        time.Time      `json:"updated"`
	MembersUpdated time.Time      `db:"members_updated" json:"members_updated"`
}

// MarshalJSON encoding this to json.
func (d *Chat) MarshalJSON() ([]byte, error) {
	type Alias Chat
	return json.Marshal(&struct {
		*Alias
		ID      string `json:"id"`
		Owner   string `json:"owner,omitempty"`
		Created int64  `json:"created"`
		Updated int64  `json:"updated"`
	}{
		Alias:   (*Alias)(d),
		ID:      EncodeChatIdentity(d.Type, d.ID),
		Owner:   d.Owner.String,
		Created: d.Created.Unix(),
		Updated: d.Updated.Unix(),
	})
}

// RoomChat is a room conversation.
type RoomChat struct {
	Area   uint32 `db:"area" json:"area"`
	ChatID uint64 `db:"chat_id" json:"chat_id"`
}

// UserChat is a user's conversation.
type UserChat struct {
	ID             uint64    `db:"id" json:"id"`
	Type           string    `json:"type"`
	Title          string    `json:"title"`
	Tag            string    `json:"tag"`
	MsgID          uint64    `db:"msg_id" json:"msg_id"`
	Ext            string    `db:"ext" json:"ext"`
	Created        time.Time `json:"created"`
	Updated        time.Time `json:"updated"`
	MembersUpdated time.Time `db:"members_updated" json:"members_updated"`
	User           string    `json:"user"`
	CurID          uint64    `db:"cur_id" json:"cur_id"`
	Joined         time.Time `json:"joined"`
	ExitMsgID      uint64    `db:"exit_msg_id" json:"exit_msg_id"`
	IsExited       bool      `db:"is_exited" json:"is_exited"`
	Dnd            bool      `json:"dnd"`
	Label          string    `json:"label"`
	JoinMsgID      uint64    `db:"join_msg_id" json:"join_msg_id"`
	LastMsgTs      time.Time `db:"last_msg_ts" json:"last_msg_ts"`
	UserUpdated    time.Time `db:"user_updated" json:"user_updated"`
}

// MarshalJSON encoding this to json.
func (d *UserChat) MarshalJSON() ([]byte, error) {
	type Alias UserChat
	return json.Marshal(&struct {
		*Alias
		ID             string `json:"id"`
		Created        int64  `json:"created"`
		Updated        int64  `json:"updated"`
		Joined         int64  `json:"joined"`
		LastMsgTs      int64  `json:"last_msg_ts"`
		MembersUpdated int64  `json:"members_updated"`
		UserUpdated    int64  `json:"user_updated"`
	}{
		Alias:          (*Alias)(d),
		ID:             EncodeChatIdentity(d.Type, d.ID),
		Created:        d.Created.Unix(),
		Updated:        d.Updated.Unix(),
		Joined:         d.Joined.Unix(),
		LastMsgTs:      d.LastMsgTs.Unix(),
		MembersUpdated: d.MembersUpdated.Unix(),
		UserUpdated:    d.UserUpdated.Unix(),
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
	Label     string    `db:"label"`
	Updated   time.Time `db:"updated"`
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
