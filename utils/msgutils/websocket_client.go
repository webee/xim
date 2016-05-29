package msgutils

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

// WSClientTranseiver is a client websocket connection.
type WSClientTranseiver struct {
	ws             *websocket.Conn
	serializer     Serializer
	messages       chan Message
	toSendMessages chan Message
	closed         bool
}

// NewWSClientTranseiver creates a new client websocket Transeiver.
func NewWSClientTranseiver(ws *websocket.Conn, serializer Serializer, chanSize int) *WSClientTranseiver {
	t := &WSClientTranseiver{
		ws:             ws,
		serializer:     serializer,
		messages:       make(chan Message, 10*chanSize),
		toSendMessages: make(chan Message, chanSize),
	}
	t.run()
	return t
}

// Run starts transeiver run loops.
func (c *WSClientTranseiver) run() {
	go c.readPump()
	go c.writePump()
}

// Receive get receive message channel.
func (c *WSClientTranseiver) Receive() <-chan Message {
	return c.messages
}

// Send get send message channel.
func (c *WSClientTranseiver) Send(msg Message) error {
	if c.closed {
		return TranseiverError("closed")
	}
	c.toSendMessages <- msg
	return nil
}

func (c *WSClientTranseiver) write(mt int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
	return c.ws.WriteMessage(mt, payload)
}

// Close closes the underlying websocket connection.
func (c *WSClientTranseiver) Close() error {
	if !c.closed {
		c.closed = true
		close(c.toSendMessages)
		log.Println("websocket transeiver closed.")
		return c.ws.Close()
	}
	return nil
}

func (c *WSClientTranseiver) writePump() {
	defer func() {
		c.ws.Close()
	}()

	for {
		msg, ok := <-c.toSendMessages
		if !ok {
			c.write(websocket.CloseMessage, []byte{})
			return
		}
		b, err := c.serializer.Serialize(msg)
		if err != nil {
			log.Println("error serializing message:", err)
			// TODO: handle error
			continue
		}
		if err := c.write(websocket.TextMessage, b); err != nil {
			return
		}
	}
}

func (c *WSClientTranseiver) readPump() {
	defer func() {
		close(c.messages)
		c.ws.Close()
	}()

	for {
		msgType, b, err := c.ws.ReadMessage()

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("ws error:%v", err)
			}
			break
		} else if msgType == websocket.CloseMessage {
			break
		}
		msg, err := c.serializer.Deserialize(b)
		if err != nil {
			log.Println("error deserializing message:", err)
			// TODO: handle error
			continue
		}
		c.messages <- msg
	}
}
