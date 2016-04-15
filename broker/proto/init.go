package proto

import "encoding/json"

const (
	HelloMsg   = "hello"
	WelcomeMsg = "welcome"
	PingMsg    = "ping"
	PongMsg    = "pong"
	ByeMsg     = "bye"
	MsgMsg     = "msg"
)

// BaseMsg is the base msg type.
type Base struct {
	Type string `json:"type"`
}

// Hello is the auth msg.
type Hello struct {
	Base
	Token string `json:"token"`
}

// Welcome is the success authentication msg.
type Welcome struct {
	Base
}

// Ping is the heartbeat msg.
type Ping struct {
	Base
}

// Pong is the heartbeat reply msg.
type Pong struct {
	Base
}

// Bye is the msg before close.
type Bye struct {
	Base
}

// Msg is the real message msg.
type Msg struct {
	Base
	Msg json.RawMessage `json:"msg"`
}

// MsgType return the base msg type.
func (b *Base) MsgType() string {
	return b.Type
}

// NewWelcome create a welcome msg.
func NewWelcome() *Welcome {
	return &Welcome{Base{Type: HelloMsg}}
}

// NewBye create a bye msg.
func NewBye() *Bye {
	return &Bye{Base{Type: ByeMsg}}
}

// NewPong create a pong msg.
func NewPong() *Pong {
	return &Pong{Base{Type: PongMsg}}
}
