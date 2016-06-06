package mid

import (
	"fmt"
	"time"
	pubtypes "xim/xchat/logic/pub/types"
)

// Timestamp is timestamp.
type Timestamp time.Time

// MarshalJSON encode Timestamp to byte array.
func (t Timestamp) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%d", time.Time(t).Unix())), nil
}

// Chat is a chat.
type Chat struct {
	ID    uint64 `json:"id"`
	Type  string `json:"type"`
	Tag   string `json:"tag"`
	Title string `json:"title"`
}

// Message is a chat message.
type Message struct {
	User string `json:"user"`
	ID   uint64 `json:"id"`
	Ts   int64  `json:"ts"`
	Msg  string `json:"msg"`
}

// ChatMessages is chat's many messages.
type ChatMessages struct {
	ChatID uint64     `json:"chat_id"`
	Type   string     `json:"type"`
	Tag    string     `json:"tag"`
	Title  string     `json:"title"`
	Msgs   []*Message `json:"msgs"`
}

// NewMessageFromDBMsg converts db.Message to Message.
func NewMessageFromDBMsg(msg *pubtypes.Message) *Message {
	return &Message{
		User: msg.User,
		ID:   msg.ID,
		Ts:   msg.Ts,
		Msg:  msg.Msg,
	}
}
