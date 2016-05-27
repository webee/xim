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
	conn       *websocket.Conn
	serializer Serializer
	messages   chan Message
	closed     bool
	sendMutex  sync.Mutex
}

// NewWSTranseiver creates a new websocket Transeiver.
func NewWSTranseiver(conn *websocket.Conn, serializer Serializer, chanSize int) *WSTranseiver {
	c := &WSTranseiver{
		conn:       conn,
		serializer: serializer,
		messages:   make(chan Message, chanSize),
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
	c.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	defer c.conn.SetWriteDeadline(time.Time{})

	err = c.conn.WriteMessage(websocket.TextMessage, b)
	if err != nil {
		c.Close()
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

func (c *WSTranseiver) run() {
	defer func() {
		if r := recover(); r != nil {
			log.Println("transeiver panic:", r)
		}
	}()

	var msgType int
	var b []byte
	var err error
	for {
		msgType, b, err = c.conn.ReadMessage()

		if err != nil {
			log.Println("read error:", err)
			c.Close()
			break
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
