package mid

import (
	"fmt"
	"time"
	"xim/xchat/logic/db"
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
	ID      uint64    `json:"id"`
	Type    string    `json:"type"`
	Tag     string    `json:"tag"`
	Title   string    `json:"title"`
	MsgID   uint64    `json:"msg_id"`
	Created Timestamp `json:"created"`
}

// NewChatFromDBChat converts db.Chat to Chat.
func NewChatFromDBChat(c *db.Chat) *Chat {
	return &Chat{
		ID:      c.ID,
		Type:    c.Type,
		Tag:     c.Tag,
		Title:   c.Title,
		MsgID:   c.MsgID,
		Created: Timestamp(c.Created),
	}
}

// Message is a chat message.
type Message struct {
	ChatID uint64 `json:"chat_id"`
	User   string `json:"user"`
	ID     uint64 `json:"id"`
	Ts     int64  `json:"ts"`
	Msg    string `json:"msg"`
}

// NewMessageFromDBMsg converts db.Message to Message.
func NewMessageFromDBMsg(msg *pubtypes.ChatMessage) *Message {
	return &Message{
		ChatID: msg.ChatID,
		User:   msg.User,
		ID:     msg.ID,
		Ts:     msg.Ts,
		Msg:    msg.Msg,
	}
}

// NotifyMessage is a chat notify message.
type NotifyMessage struct {
	ChatID uint64 `json:"chat_id"`
	User   string `json:"user"`
	Ts     int64  `json:"ts"`
	Msg    string `json:"msg"`
}

// NewNotifyMessageFromDBMsg converts db.Message to Message.
func NewNotifyMessageFromDBMsg(msg *pubtypes.ChatNotifyMessage) *NotifyMessage {
	return &NotifyMessage{
		ChatID: msg.ChatID,
		User:   msg.User,
		Ts:     msg.Ts,
		Msg:    msg.Msg,
	}
}
