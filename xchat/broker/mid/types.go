package mid

import (
	"fmt"
	pubtypes "xim/xchat/logic/pub/types"
	"xim/xchat/logic/service/types"
)

// StatefulMsg is a stateful message.
type StatefulMsg interface {
	Kind() string
	State() interface{}
}

// StatelessMsg is a stateless message.
type StatelessMsg interface {
	Kind() string
}

// Message is a chat message.
type Message struct {
	ChatID string `json:"chat_id"`
	User   string `json:"user"`
	ID     uint64 `json:"id"`
	Ts     int64  `json:"ts"`
	Msg    string `json:"msg"`
}

// Kind returns the message kind.
func (msg Message) Kind() string {
	return types.MsgKindChat
}

// State get stateful message's state.
func (msg Message) State() interface{} {
	return msg.ID
}

// NewMessageFromPubMsg converts pubtypes.ChatMessage to Message.
func NewMessageFromPubMsg(msg *pubtypes.ChatMessage) *Message {
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

// Kind returns the message kind.
func (msg NotifyMessage) Kind() string {
	return types.MsgKindChatNotify
}

// NewNotifyMessageFromPubMsg converts pubtypes.ChatNotifyMessage to NotifyMessage.
func NewNotifyMessageFromPubMsg(msg *pubtypes.ChatNotifyMessage) *NotifyMessage {
	return &NotifyMessage{
		ChatID: fmt.Sprintf("%s.%d", msg.ChatType, msg.ChatID),
		User:   msg.User,
		Ts:     msg.Ts,
		Msg:    msg.Msg,
	}
}

// UserNotifyMessage is a user notify message.
type UserNotifyMessage struct {
	Ts  int64  `json:"ts"`
	Msg string `json:"msg"`
}

// Kind returns the message kind.
func (msg UserNotifyMessage) Kind() string {
	return types.MsgKindUserNotify
}

// NewUserNotifyMessageFromPubMsg converts pubtypes.UserNotifyMessage to UserNotifyMessage.
func NewUserNotifyMessageFromPubMsg(msg *pubtypes.UserNotifyMessage) *UserNotifyMessage {
	return &UserNotifyMessage{
		Ts:  msg.Ts,
		Msg: msg.Msg,
	}
}

// RawMessage is a raw message.
type RawMessage struct {
	Kind string
	Msgs []interface{}
}
