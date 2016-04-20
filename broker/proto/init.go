package proto

import "encoding/json"

// msg types.
const (
	HelloMsg   = "hello"
	WelcomeMsg = "welcome"
	PingMsg    = "ping"
	PongMsg    = "pong"
	ByeMsg     = "bye"
	ErrorReply = "error"
)

// Msg is the base user send msg.
type Msg struct {
	ID   int             `json:"id"`
	Type string          `json:"type,omitempty"`
	Msg  json.RawMessage `json:"msg,omitempty"`
}

// MsgWithBytes is msg with full msg bytes.
type MsgWithBytes struct {
	Msg
	Bytes []byte
}

// Reply is the base server reply msg.
type Reply struct {
	ReplyTo int             `json:"reply_to"`
	Type    string          `json:"type,omitempty"`
	Msg     json.RawMessage `json:"msg,omitempty"`
	Err     string          `json:"err,omitempty"`
}

// Hello is the auth msg.
type Hello struct {
	Msg
	Token string `json:"token"`
}

// MsgMsg is msg msg.
type MsgMsg struct {
	Type    string          `json:"type"`
	Channel string          `json:"channel"`
	ID      string          `json:"id"`
	LastID  string          `json:"last_id"`
	User    string          `json:"user"`
	Msg     json.RawMessage `json:"msg"`
}

// NewReply create a reply msg.
func NewReply(replyTo int, msgType string, msg json.RawMessage) *Reply {
	return &Reply{ReplyTo: replyTo, Type: msgType, Msg: msg}
}

// NewWelcome create a welcome msg.
func NewWelcome(replyTo int) *Reply {
	return NewReply(replyTo, WelcomeMsg, nil)
}

// NewPong create a pong msg.
func NewPong(replyTo int) *Reply {
	return NewReply(replyTo, PongMsg, nil)
}

// NewReplyBye create a reply bye msg.
func NewReplyBye(replyTo int) *Reply {
	return NewReply(replyTo, ByeMsg, nil)
}

// NewErrorReply create a error reply msg.
func NewErrorReply(replyTo int, err string) *Reply {
	return &Reply{ReplyTo: replyTo, Type: ErrorReply, Err: err}
}
