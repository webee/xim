package proto

// msg types.
const (
	HelloMsg = "hello"
	PingMsg  = "ping"
	PongMsg  = "pong"
	ByeMsg   = "bye"

	ReplyMsg     = "reply"
	PutMsg       = "put"
	PutStatusMsg = "status"
	PutNotifyMsg = "notify"
	MsgMsg       = "msg"

	AppNullMsg           = "null"
	AppRegisterUserMsg   = "reg"
	AppUnregisterUserMsg = "unreg"
)

// Msg is the base user send msg.
type Msg struct {
	UID     uint32      `json:"uid"`
	User    string      `json:"user,omitempty"`
	SN      interface{} `json:"sn"`
	Type    string      `json:"type,omitempty"`
	Channel string      `json:"channel,omitempty"`
	Kind    string      `json:"kind,omitempty"`
	Msg     interface{} `json:"msg,omitempty"`
}

// Reply is the base server reply msg.
type Reply struct {
	UID  interface{} `json:"uid,omitempty"`
	User string      `json:"user,omitempty"`
	SN   interface{} `json:"sn,omitempty"`
	Ok   interface{} `json:"ok,omitempty"`
	Type string      `json:"type,omitempty"`
	Data interface{} `json:"data,omitempty"`
	Err  string      `json:"err,omitempty"`
}

// TypeMsg is a type with a msg payload.
type TypeMsg struct {
	UID  interface{} `json:"uid,omitempty"`
	Type string      `json:"type,omitempty"`
	Msg  interface{} `json:"msg,omitempty"`
}

// ChannelMsg is channel event msg.
type ChannelMsg struct {
	UID       interface{} `json:"uid,omitempty"`
	Type      string      `json:"type,omitempty"`
	ID        interface{} `json:"id,omitempty"`
	User      string      `json:"user"`
	Channel   string      `json:"channel"`
	Timestamp interface{} `json:"ts"`
	Kind      string      `json:"kind,omitempty"`
	Msg       interface{} `json:"msg"`
}

// NewHello create a hello msg.
func NewHello() *TypeMsg {
	return &TypeMsg{Type: HelloMsg}
}

// NewPing create a ping msg.
func NewPing(msg interface{}) *TypeMsg {
	return &TypeMsg{Type: PongMsg, Msg: msg}
}

// NewPong create a pong msg.
func NewPong(msg interface{}) *TypeMsg {
	return &TypeMsg{Type: PongMsg, Msg: msg}
}

// NewBye create a bye msg.
func NewBye() *TypeMsg {
	return &TypeMsg{Type: ByeMsg}
}

// NewReplyRegister create a reply bye msg.
func NewReplyRegister(replyTo interface{}, user string, uid uint32) *Reply {
	return &Reply{Type: ReplyMsg, SN: replyTo, Ok: true, User: user, UID: uid}
}

// NewReplyUnregister create a reply bye msg.
func NewReplyUnregister(replyTo interface{}, uid uint32) *Reply {
	return &Reply{Type: ReplyMsg, SN: replyTo, Ok: true, UID: uid}
}

// NewReply create a reply msg.
func NewReply(replyTo interface{}, data interface{}) *Reply {
	return &Reply{Type: ReplyMsg, SN: replyTo, Ok: true, Data: data}
}

// NewErrorReply create a error reply msg.
func NewErrorReply(replyTo interface{}, err string) *Reply {
	return &Reply{Type: ReplyMsg, SN: replyTo, Ok: false, Err: err}
}
