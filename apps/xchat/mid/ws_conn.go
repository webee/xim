package mid

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

// WsConn is a websocket connection.
type wsConnection struct {
	conn       *websocket.Conn
	serializer Serializer
	messages   chan map[string]interface{}
	closed     bool
	sendMutex  sync.Mutex
}

func newWsConnection(conn *websocket.Conn, chanSize int) *wsConnection {
	c := &wsConnection{
		conn:       conn,
		serializer: new(JSONSerializer),
		messages:   make(chan map[string]interface{}, chanSize),
	}
	go c.run()
	return c
}

// TODO: make this just add the message to a channel so we don't block
func (c *wsConnection) Send(msg interface{}) error {
	b, err := c.serializer.Serialize(msg)
	if err != nil {
		return err
	}
	c.sendMutex.Lock()
	defer c.sendMutex.Unlock()

	return c.conn.WriteMessage(websocket.TextMessage, b)
}

func (c *wsConnection) Receive() <-chan map[string]interface{} {
	return c.messages
}

// Close closes the underlying websocket connection.
func (c *wsConnection) Close() error {
	return c.conn.Close()
}

func (c *wsConnection) run() {
	for {
		// NOTE: read timeout(heartbeat.)
		msgType, b, err := c.conn.ReadMessage()
		if err != nil {
			if c.closed {
				log.Println("peer connection closed")
			} else {
				log.Println("error reading from peer:", err)
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
				log.Println("error deserializing peer message:", err)
				// TODO: handle error
			} else {
				c.messages <- msg
			}
		}
	}
}
