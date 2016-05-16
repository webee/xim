package ws

import (
	"fmt"
	"time"
	"xim/broker/proto"
)

// Sender can send messages.
type Sender interface {
	Send(msg interface{}) error
}

// Connection is a client websocket connection.
type Connection interface {
	Sender
	Close() error
	Receive() <-chan *proto.Msg
}

// GetMessageTimeout is a convenience function to get a single message from a
// connection within a specified period of time
func GetMessageTimeout(c Connection, t time.Duration) (interface{}, error) {
	select {
	case msg, open := <-c.Receive():
		if !open {
			return nil, fmt.Errorf("receive channel closed")
		}
		return msg, nil
	case <-time.After(t):
		return nil, fmt.Errorf("timeout waiting for message")
	}
}
