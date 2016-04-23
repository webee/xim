package proto

// msg types.
const (
	HelloMsg   = "hello"
	WelcomeMsg = "welcome"
	PingMsg    = "ping"
	PongMsg    = "pong"
	ByeMsg     = "bye"
	RespReply  = "resp"
	ErrorReply = "error"
)

// Msg is the base user send msg.
type Msg struct {
	ID      int         `json:"id"`
	Type    string      `json:"type,omitempty"`
	Channel string      `json:"channel,omitempty"`
	Msg     interface{} `json:"msg,omitempty"`
}

// MsgWithBytes is msg with full msg bytes.
type MsgWithBytes struct {
	Msg
	Bytes []byte
}

// Reply is the base server reply msg.
type Reply struct {
	ReplyTo int         `json:"reply_to"`
	Type    string      `json:"type,omitempty"`
	Msg     interface{} `json:"msg,omitempty"`
	Err     string      `json:"err,omitempty"`
}

// MsgMsg is msg msg.
type MsgMsg struct {
	Channel string      `json:"channel"`
	ID      string      `json:"id"`
	LastID  string      `json:"last_id"`
	User    string      `json:"user"`
	Msg     interface{} `json:"msg"`
}

// NewReply create a reply msg.
func NewReply(replyTo int, msgType string, msg interface{}) *Reply {
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

// NewResponse create a response msg.
func NewResponse(replyTo int, msg interface{}) *Reply {
	return NewReply(replyTo, RespReply, msg)
}

// NewErrorReply create a error reply msg.
func NewErrorReply(replyTo int, err string) *Reply {
	return &Reply{ReplyTo: replyTo, Type: ErrorReply, Err: err}
}
