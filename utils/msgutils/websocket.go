package msgutils

import (
	"errors"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WSTranseiver is a websocket connection.
type WSTranseiver struct {
	sync.Mutex
	conn             *websocket.Conn
	serializer       Serializer
	messages         chan Message
	closed           bool
	sendMutex        sync.Mutex
	heartbeatTimeout time.Duration
}

// NewWSTranseiver creates a new websocket Transeiver.
func NewWSTranseiver(conn *websocket.Conn, serializer Serializer, chanSize int, heartbeatTimeout time.Duration) *WSTranseiver {
	c := &WSTranseiver{
		conn:             conn,
		serializer:       serializer,
		messages:         make(chan Message, chanSize),
		heartbeatTimeout: heartbeatTimeout,
	}
	go c.run()
	return c
}

// Send send messages.
func (c *WSTranseiver) Send(msg Message) error {
	if c.closed {
		return errors.New("transeiver closed")
	}

	b, err := c.serializer.Serialize(msg)
	if err != nil {
		return err
	}
	c.sendMutex.Lock()
	defer c.sendMutex.Unlock()

	// NOTE: send timeout.
	c.conn.SetWriteDeadline(time.Now().Add(7 * time.Second))
	defer c.conn.SetWriteDeadline(time.Time{})

	err = c.conn.WriteMessage(websocket.TextMessage, b)
	if IsCloseError(err) {
		c.conn.Close()
	}

	return err
}

// Receive get message channel.
func (c *WSTranseiver) Receive() <-chan Message {
	return c.messages
}

// Close closes the underlying websocket connection.
func (c *WSTranseiver) Close() error {
	if !c.closed {
		c.Lock()
		defer c.Unlock()
		close(c.messages)
		c.closed = true
		log.Println("websocket transeiver closed.")
		return c.conn.Close()
	}
	return nil
}

// Closed return wheather the transeiver is closed or not.
func (c *WSTranseiver) Closed() bool {
	return c.closed
}

// IsCloseError checks if the error is a websocket closed error.
func IsCloseError(err error) bool {
	_, ok := err.(*websocket.CloseError)
	return ok
}

func (c *WSTranseiver) run() {
	var msgType int
	var b []byte
	var err error
	for {
		// NOTE: read timeout(heartbeat.)
		if c.heartbeatTimeout > 0 {
			c.conn.SetReadDeadline(time.Now().Add(c.heartbeatTimeout))
			msgType, b, err = c.conn.ReadMessage()
			c.conn.SetReadDeadline(time.Time{})
		} else {
			msgType, b, err = c.conn.ReadMessage()
		}

		if err != nil {
			if IsCloseError(err) {
				c.Close()
				break
			}
			log.Println("read error:", err)
		} else if msgType == websocket.CloseMessage {
			c.Close()
			break
		} else {
			msg, err := c.serializer.Deserialize(b)
			if err != nil {
				log.Println("error deserializing message:", err)
				// TODO: handle error
			} else {
				c.messages <- msg
			}
		}
	}
}
