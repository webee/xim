package mid

import (
	"xim/xchat/logic/db"
	pubtypes "xim/xchat/logic/pub/types"
	"xim/xchat/logic/service/types"
)

// Msg is a kind message.
type Msg interface {
	Kind() string
}

// StatefulMsg is a stateful message.
type StatefulMsg interface {
	Msg
	State() interface{}
}

// StatelessMsg is a stateless message.
type StatelessMsg interface {
	Msg
}

// Message is a chat message.
type Message struct {
	ChatID string `json:"chat_id"`
	Domain string `json:"domain,omitempty"`
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
		ChatID: db.EncodeChatIdentity(msg.ChatType, msg.ChatID),
		Domain: msg.Domain,
		User:   msg.User,
		ID:     msg.ID,
		Ts:     msg.Ts,
		Msg:    msg.Msg,
	}
}

// NotifyMessage is a chat notify message.
type NotifyMessage struct {
	ChatID string `json:"chat_id"`
	Domain string `json:"domain,omitempty"`
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
		ChatID: db.EncodeChatIdentity(msg.ChatType, msg.ChatID),
		Domain: msg.Domain,
		User:   msg.User,
		Ts:     msg.Ts,
		Msg:    msg.Msg,
	}
}

// UserNotifyMessage is a user notify message.
type UserNotifyMessage struct {
	Domain string `json:"domain,omitempty"`
	Ts     int64  `json:"ts"`
	Msg    string `json:"msg"`
}

// Kind returns the message kind.
func (msg UserNotifyMessage) Kind() string {
	return types.MsgKindUserNotify
}

// NewUserNotifyMessageFromPubMsg converts pubtypes.UserNotifyMessage to UserNotifyMessage.
func NewUserNotifyMessageFromPubMsg(msg *pubtypes.UserNotifyMessage) *UserNotifyMessage {
	return &UserNotifyMessage{
		Domain: msg.Domain,
		Ts:     msg.Ts,
		Msg:    msg.Msg,
	}
}

// RawMessage is a raw message.
type RawMessage struct {
	Kind string
	Msgs []interface{}
}
