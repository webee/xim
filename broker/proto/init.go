package proto

// msg types.
const (
	HelloMsg   = "hello"
	WelcomeMsg = "welcome"
	PingMsg    = "ping"
	PongMsg    = "pong"
	ByeMsg     = "bye"
)

// Msg is the base user send msg.
type Msg struct {
	ID   int    `json:"id"`
	Type string `json:"type"`
}

// MsgWithBytes is msg with full msg bytes.
type MsgWithBytes struct {
	Msg
	Bytes []byte
}

// Reply is the base server reply msg.
type Reply struct {
	ReplyTo int    `json:"reply_to"`
	Type    string `json:"type"`
}

// Hello is the auth msg.
type Hello struct {
	Msg
	Token string `json:"token"`
}

// Welcome is the success authentication msg.
type Welcome struct {
	Reply
}

// Ping is the heartbeat msg.
type Ping struct {
	Msg
}

// Pong is the heartbeat reply msg.
type Pong struct {
	Reply
}

// Bye is the msg before close.
type Bye struct {
	Msg
}

// ReplyBye is the reply msg for bye.
type ReplyBye struct {
	Reply
}

// ReplyError is the reply msg when error.
type ReplyError struct {
	Reply
	Ok     bool   `json:"ok"`
	Reason string `json:"reason"`
}

// NewWelcome create a welcome msg.
func NewWelcome(replyTo int) *Welcome {
	return &Welcome{Reply{ReplyTo: replyTo, Type: WelcomeMsg}}
}

// NewPong create a pong msg.
func NewPong(replyTo int) *Pong {
	return &Pong{Reply{ReplyTo: replyTo, Type: PongMsg}}
}

// NewReplyBye create a reply bye msg.
func NewReplyBye(replyTo int) *ReplyBye {
	return &ReplyBye{Reply{ReplyTo: replyTo, Type: ByeMsg}}
}

// NewReplyError create a reply error msg.
func NewReplyError(replyTo int, t, reason string) *ReplyError {
	return &ReplyError{Reply{ReplyTo: replyTo, Type: t}, false, reason}
}
