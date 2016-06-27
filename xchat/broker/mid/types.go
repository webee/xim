package mid

import (
	"fmt"
	pubtypes "xim/xchat/logic/pub/types"
)

// Message is a chat message.
type Message struct {
	ChatID string `json:"chat_id"`
	User   string `json:"user"`
	ID     uint64 `json:"id"`
	Ts     int64  `json:"ts"`
	Msg    string `json:"msg"`
}

// NewMessageFromDBMsg converts db.Message to Message.
func NewMessageFromDBMsg(msg *pubtypes.ChatMessage) *Message {
	return &Message{
		ChatID: fmt.Sprintf("%s.%d", msg.ChatType, msg.ChatID),
		User:   msg.User,
		ID:     msg.ID,
		Ts:     msg.Ts,
		Msg:    msg.Msg,
	}
}

// NotifyMessage is a chat notify message.
type NotifyMessage struct {
	ChatID string `json:"chat_id"`
	User   string `json:"user"`
	Ts     int64  `json:"ts"`
	Msg    string `json:"msg"`
}

// NewNotifyMessageFromDBMsg converts db.Message to Message.
func NewNotifyMessageFromDBMsg(msg *pubtypes.ChatNotifyMessage) *NotifyMessage {
	return &NotifyMessage{
		ChatID: fmt.Sprintf("%s.%d", msg.ChatType, msg.ChatID),
		User:   msg.User,
		Ts:     msg.Ts,
		Msg:    msg.Msg,
	}
}
