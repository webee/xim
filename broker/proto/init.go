package proto

// msg types.
const (
	HelloMsg     = "hello"
	PingMsg      = "ping"
	PongMsg      = "pong"
	ByeMsg       = "bye"
	RespReply    = "resp"
	ErrorReply   = "error"
	PutMsg       = "put"
	PutStatusMsg = "status"
	PutNotifyMsg = "notify"
	MsgMsg       = "msg"

	AppNullMsg           = "null"
	AppRegisterUserMsg   = "register"
	AppUnregisterUserMsg = "unregister"
)

// Msg is the base user send msg.
type Msg struct {
	UID     uint32      `json:"uid"`
	User    string      `json:"user,omitempty"`
	ID      int         `json:"id"`
	Type    string      `json:"type,omitempty"`
	Channel string      `json:"channel,omitempty"`
	Kind    string      `json:"kind,omitempty"`
	Msg     interface{} `json:"msg,omitempty"`
}

// MsgWithBytes is msg with full msg bytes.
type MsgWithBytes struct {
	Msg
	Bytes []byte
}

// Reply is the base server reply msg.
type Reply struct {
	UID     interface{} `json:"uid,omitempty"`
	User    string      `json:"user,omitempty"`
	ReplyTo interface{} `json:"reply_to,omitempty"`
	Ok      interface{} `json:"ok,omitempty"`
	Type    string      `json:"type,omitempty"`
	Msg     interface{} `json:"msg,omitempty"`
	Err     string      `json:"err,omitempty"`
}

// ChannelMsg is channel event msg.
type ChannelMsg struct {
	UID     interface{} `json:"uid,omitempty"`
	Type    string      `json:"type,omitempty"`
	ID      string      `json:"id,omitempty"`
	LastID  string      `json:"last_id,omitempty"`
	User    string      `json:"user"`
	Channel string      `json:"channel"`
	Kind    string      `json:"kind,omitempty"`
	Msg     interface{} `json:"msg"`
}

// NewReply create a reply msg.
func NewReply(replyTo interface{}, msgType string, msg interface{}) *Reply {
	return &Reply{ReplyTo: replyTo, Type: msgType, Msg: msg}
}

// NewHello create a hello msg.
func NewHello() *Reply {
	return NewReply(nil, HelloMsg, nil)
}

// NewPong create a pong msg.
func NewPong(replyTo int, msg interface{}) *Reply {
	return NewReply(replyTo, PongMsg, msg)
}

// NewReplyBye create a reply bye msg.
func NewReplyBye(replyTo int) *Reply {
	return NewReply(replyTo, ByeMsg, nil)
}

// NewReplyRegister create a reply bye msg.
func NewReplyRegister(replyTo int, user string, uid uint32) *Reply {
	return &Reply{Ok: true, ReplyTo: replyTo, Type: AppRegisterUserMsg, User: user, UID: uid}
}

// NewReplyUnregister create a reply bye msg.
func NewReplyUnregister(replyTo int, uid uint32) *Reply {
	return &Reply{Ok: true, ReplyTo: replyTo, Type: AppUnregisterUserMsg, UID: uid}
}

// NewResponse create a response msg.
func NewResponse(replyTo int, msg interface{}) *Reply {
	return &Reply{ReplyTo: replyTo, Ok: true, Type: RespReply, Msg: msg}
}

// NewErrorReply create a error reply msg.
func NewErrorReply(replyTo int, err string) *Reply {
	return &Reply{ReplyTo: replyTo, Ok: false, Type: ErrorReply, Err: err}
}
