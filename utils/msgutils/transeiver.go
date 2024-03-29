package msgutils

import (
	"fmt"
	"time"
)

// TranseiverError is transeiver errors.
type TranseiverError string

func (t TranseiverError) Error() string {
	return string(t)
}

// A Sender can send a message to its peer.
//
// For clients, this sends a message to the router, and for routers,
// this sends a message to the client.
type Sender interface {
	// Send a message to the peer
	Send(Message) error
}

// Transeiver is the interface that must be implemented by all WAMP peers.
type Transeiver interface {
	Sender

	// Closes the peer connection and any channel returned from Receive().
	// Multiple calls to Close() will have no effect.
	Close() error

	// Receive returns a channel of messages coming from the peer.
	Receive() <-chan Message
}

// GetMessageTimeout is a convenience function to get a single message from a
// peer within a specified period of time
func GetMessageTimeout(p Transeiver, t time.Duration) (Message, error) {
	select {
	case msg, open := <-p.Receive():
		if !open {
			return nil, fmt.Errorf("receive channel closed")
		}
		return msg, nil
	case <-time.After(t):
		return nil, fmt.Errorf("timeout waiting for message")
	}
}
