package msgutils

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

// WSTranseiver is a websocket connection.
type WSTranseiver struct {
	ws             *websocket.Conn
	serializer     Serializer
	messages       chan Message
	toSendMessages chan Message
	closed         bool
	pongWait       time.Duration
	pongMsg        Message
	pingPeriod     time.Duration
}

// NewWSTranseiver creates a new websocket Transeiver.
func NewWSTranseiver(ws *websocket.Conn, serializer Serializer, chanSize int, pongWait time.Duration, pongMsg Message) *WSTranseiver {
	t := &WSTranseiver{
		ws:             ws,
		serializer:     serializer,
		messages:       make(chan Message, chanSize),
		toSendMessages: make(chan Message, 10*chanSize),
		pongWait:       pongWait,
		pongMsg:        pongMsg,
		pingPeriod:     (pongWait * 9) / 10,
	}
	t.run()
	return t
}

// Run starts transeiver run loops.
func (c *WSTranseiver) run() {
	go c.readPump()
	go c.writePump()
}

// Receive get receive message channel.
func (c *WSTranseiver) Receive() <-chan Message {
	return c.messages
}

// Send get send message channel.
func (c *WSTranseiver) Send(msg Message) (err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("panic:", r)
			err = TranseiverError("closed")
		}
	}()

	if c.closed {
		return TranseiverError("closed")
	}
	c.toSendMessages <- msg
	return nil
}

func (c *WSTranseiver) write(mt int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
	return c.ws.WriteMessage(mt, payload)
}

// Close closes the underlying websocket connection.
func (c *WSTranseiver) Close() error {
	if !c.closed {
		c.closed = true
		close(c.toSendMessages)
		log.Println("websocket transeiver closed.")
		return c.ws.Close()
	}
	return nil
}

func (c *WSTranseiver) writePump() {
	ticker := time.NewTicker(c.pingPeriod)
	defer func() {
		ticker.Stop()
		c.ws.Close()
	}()

	for {
		select {
		case msg, ok := <-c.toSendMessages:
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
		case <-ticker.C:
			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

func (c *WSTranseiver) readPump() {
	defer func() {
		close(c.messages)
		c.ws.Close()
	}()

	// NOTE: limit to 16Kb.
	c.ws.SetReadLimit(16 * 1024)
	c.ws.SetReadDeadline(time.Now().Add(c.pongWait))
	c.ws.SetPongHandler(func(string) error {
		c.ws.SetReadDeadline(time.Now().Add(c.pongWait))
		if c.pongMsg != nil {
			c.messages <- c.pongMsg
		}
		return nil
	})

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
