package proto

import (
	"xim/utils/msgutils"
)

// put message kinds.
const (
	PutStatusMsgKind = "status"
	PutNotifyMsgKind = "notify"
)

// XIMMsgType is XIM MessageType type.
type XIMMsgType msgutils.MessageType

// New creates a message.
func (mt XIMMsgType) New() msgutils.Message {
	switch mt {
	case NULL:
		return new(Null)
	case HELLO:
		return new(Hello)
	case PING:
		return new(Ping)
	case PONG:
		return new(Pong)
	case BYE:
		return new(Bye)
	case PUT:
		return new(Put)
	case PUSH:
		return new(Push)
	case REPLY:
		return new(Reply)
	case REGISTER:
		return new(Register)
	case UNREGISTER:
		return new(Unregister)
	default:
		return nil
	}
}

func (mt XIMMsgType) String() string {
	switch mt {
	case NULL:
		return ""
	case HELLO:
		return "hello"
	case PING:
		return "ping"
	case PONG:
		return "pong"
	case BYE:
		return "bye"
	case PUT:
		return "put"
	case PUSH:
		return "push"
	case REPLY:
		return "reply"
	case REGISTER:
		return "reg"
	case UNREGISTER:
		return "unreg"
	default:
		panic("Invalid message type")
	}
}

// Message Types.
const (
	NULL       XIMMsgType = 0 // conn: client, server
	HELLO      XIMMsgType = 1 // conn: server
	PING       XIMMsgType = 2 // conn: client
	PONG       XIMMsgType = 3 // conn: server
	BYE        XIMMsgType = 4 // conn: client, server
	PUT        XIMMsgType = 5 // user: client: req
	PUSH       XIMMsgType = 6 // user: server
	REPLY      XIMMsgType = 7 // user: server: reply
	REGISTER   XIMMsgType = 8 // conn: client<app>: req
	UNREGISTER XIMMsgType = 9 // conn: client<app>: req
)

// Null [msg]
type Null struct {
}

// MessageType returns the message type.
func (msg *Null) MessageType() msgutils.MessageType {
	return msgutils.MessageType(NULL)
}

// Hello []
type Hello struct {
}

// MessageType returns the message type.
func (msg *Hello) MessageType() msgutils.MessageType {
	return msgutils.MessageType(HELLO)
}

// Ping []
type Ping struct {
}

// MessageType returns the message type.
func (msg *Ping) MessageType() msgutils.MessageType {
	return msgutils.MessageType(PING)
}

// Pong []
type Pong struct {
}

// MessageType returns the message type.
func (msg *Pong) MessageType() msgutils.MessageType {
	return msgutils.MessageType(PONG)
}

// Bye []
type Bye struct {
}

// MessageType returns the message type.
func (msg *Bye) MessageType() msgutils.MessageType {
	return msgutils.MessageType(BYE)
}

// Put [channel, kind, msg]
type Put struct {
	SnSyncMsg `mapstructure:",squash"`
	UID       uint32      `json:"uid";mapstructure:"uid"`
	Channel   string      `json:"channel";mapstructure:"channel"`
	Kind      string      `json:"kind,omitempty";mapstructure:"kind"`
	Msg       interface{} `json:"msg";mapstructure:"msg"`
}

// MessageType returns the message type.
func (msg *Put) MessageType() msgutils.MessageType {
	return msgutils.MessageType(PUT)
}

// Push [sn, channel, user, kind, msg, id, ts]
type Push struct {
	UID     uint32      `json:"uid";mapstructure:"uid"`
	Channel string      `json:"channel"`
	User    string      `json:"user";mapstructure:"user"`
	Kind    string      `json:"kind,omitempty"`
	Msg     interface{} `json:"msg";mapstructure:"msg"`
	ID      uint64      `json:"id";mapstructure:"id"`
	Ts      uint64      `json:"ts";mapstructure:"ts"`
}

// MessageType returns the message type.
func (msg *Push) MessageType() msgutils.MessageType {
	return msgutils.MessageType(PUSH)
}

// Reply [sn, ok, <data|err_code,err_msg>]
type Reply struct {
	SnSyncMsg `mapstructure:",squash"`
	UID       uint32      `json:"uid";mapstructure:"uid"`
	Ok        bool        `jsoin:"ok";mapstructure:"ok"`
	Data      interface{} `json:"data,omitempty";mapstructure:"data"`
	ErrCode   string      `json:"err_code,omitempty";mapstructure:"err_code"`
	ErrMsg    string      `json:"err_msg,omitempty";mapstructure:"err_code"`
}

// MessageType returns the message type.
func (msg *Reply) MessageType() msgutils.MessageType {
	return msgutils.MessageType(REPLY)
}

// NewReply creates an reply
func NewReply(id msgutils.ID, data interface{}) *Reply {
	reply := &Reply{
		Ok:   true,
		Data: data,
	}
	reply.SetID(id)
	return reply
}

// NewErrorReply creates an error reply
func NewErrorReply(id msgutils.ID, errMsg string) *Reply {
	reply := &Reply{
		Ok:      false,
		ErrCode: "1",
		ErrMsg:  errMsg,
	}
	reply.SetID(id)
	return reply
}

// NewAppReply creates an app reply
func NewAppReply(id msgutils.ID, uid uint32, data interface{}) *Reply {
	reply := NewReply(id, data)
	reply.UID = uid
	return reply
}

// NewAppErrorReply creates an error reply
func NewAppErrorReply(id msgutils.ID, uid uint32, errMsg string) *Reply {
	reply := NewErrorReply(id, errMsg)
	reply.UID = uid
	return reply
}

// Register [sn, user]
type Register struct {
	SnSyncMsg `mapstructure:",squash"`
	User      string `json:"user";mapstructure:"user"`
}

// MessageType returns the message type.
func (msg *Register) MessageType() msgutils.MessageType {
	return msgutils.MessageType(REGISTER)
}

// Unregister [sn, uid]
type Unregister struct {
	SnSyncMsg `mapstructure:",squash"`
	UID       uint32 `json:"uid";mapstructure:"uid"`
}

// MessageType returns the message type.
func (msg *Unregister) MessageType() msgutils.MessageType {
	return msgutils.MessageType(UNREGISTER)
}

// SnSyncMsg is sync message with sync id sn.
type SnSyncMsg struct {
	Sn msgutils.ID `json:"sn";mapstructure:"sn"`
}

// SetID set sync message sync id.
func (msg *SnSyncMsg) SetID(sn msgutils.ID) {
	msg.Sn = sn
}

// GetID get sync message sync id.
func (msg *SnSyncMsg) GetID() msgutils.ID {
	return msg.Sn
}
