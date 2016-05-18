package msgutils

import (
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WSTranseiver is a websocket connection.
type WSTranseiver struct {
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
	b, err := c.serializer.Serialize(msg)
	if err != nil {
		return err
	}
	c.sendMutex.Lock()
	defer c.sendMutex.Unlock()

	// NOTE: send timeout.
	c.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	defer c.conn.SetWriteDeadline(time.Time{})

	return c.conn.WriteMessage(websocket.TextMessage, b)
}

// Receive get message channel.
func (c *WSTranseiver) Receive() <-chan Message {
	return c.messages
}

// Close closes the underlying websocket connection.
func (c *WSTranseiver) Close() error {
	if !c.closed {
		c.closed = true
		return c.conn.Close()
	}
	return nil
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
			if c.closed {
				log.Println("connection closed")
			} else {
				log.Println("error reading from connection:", err)
				c.conn.Close()
			}
			close(c.messages)
			break
		} else if msgType == websocket.CloseMessage {
			c.conn.Close()
			close(c.messages)
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
