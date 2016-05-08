package ws

import (
	"encoding/json"
	"errors"
	"log"
	"sync"
	"time"
	"xim/broker/proto"
	"xim/broker/userds"

	"github.com/bitly/go-simplejson"

	"github.com/gorilla/websocket"
)

// WsConn is a websocket connection.
type wsConn struct {
	wLock  sync.Mutex
	s      *WebsocketServer
	conn   *websocket.Conn
	user   *userds.UserLocation
	from   string
	msgbox chan interface{}
	done   chan struct{}
}

func newWsConn(s *WebsocketServer, user *userds.UserLocation, conn *websocket.Conn) *wsConn {
	return &wsConn{
		s:      s,
		user:   user,
		conn:   conn,
		from:   conn.RemoteAddr().String(),
		msgbox: make(chan interface{}, 5),
		done:   make(chan struct{}, 1),
	}
}

// ReadMsg read json message in a heartbeat duration.
func (c *wsConn) ReadMsg() (msg proto.Msg, err error) {
	//jd, err = c.ReadJSONData(c.s.config.HeartBeatTimeout)
	//msg.ID, err = jd.Get("id").Int()
	//msg.Type, err = jd.Get("type").String()

	// msg with bytes
	/*
		msg = new(proto.MsgWithBytes)
		bytes, err := c.ReadJSON(msg, c.s.config.HeartBeatTimeout)
		msg.Bytes = bytes
	*/
	_, err = c.ReadJSON(&msg, c.s.config.HeartbeatTimeout)
	return
}

// PushMsg write json message in a write timeout duration.
func (c *wsConn) PushMsg(user *userds.UserLocation, v interface{}) (err error) {
	select {
	case <-c.done:
		return errors.New("connection closed")
	case c.msgbox <- v:
		return nil
	}
}

// WriteMsg write json message in a write timeout duration.
func (c *wsConn) WriteMsg(v interface{}) (err error) {
	err = c.WriteJSON(v, c.s.config.WriteTimeout)
	return
}

// WriteJSON write json message in a timeout duration.
func (c *wsConn) WriteJSON(v interface{}, timeout time.Duration) error {
	conn := c.conn
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	c.wLock.Lock()
	conn.SetWriteDeadline(time.Now().Add(timeout))
	err = conn.WriteMessage(websocket.TextMessage, data)
	conn.SetWriteDeadline(time.Time{})
	c.wLock.Unlock()
	if err != nil {
		log.Println(err)
	}
	return err
}

// ReadJSON read json message in a timeout duration.
func (c *wsConn) ReadJSON(v interface{}, timeout time.Duration) (bytes []byte, err error) {
	conn := c.conn
	conn.SetReadDeadline(time.Now().Add(timeout))
	_, bytes, err = conn.ReadMessage()
	conn.SetReadDeadline(time.Time{})
	if err != nil {
		return
	}
	if len(bytes) > 16*1024 {
		err = errors.New("msg too large")
		return
	}
	/*
		if len(bytes) == 1 {
			if bytes, err = json.Marshal(map[string]string{
				"type": string(bytes),
			}); err != nil {
				return
			}
		}
	*/

	err = json.Unmarshal(bytes, v)
	return
}

// ReadJSONData read json message in a timeout duration.
func (c *wsConn) ReadJSONData(timeout time.Duration) (jd *simplejson.Json, err error) {
	conn := c.conn
	conn.SetReadDeadline(time.Now().Add(timeout))
	_, r, err := conn.NextReader()
	conn.SetReadDeadline(time.Time{})
	if err != nil {
		return
	}
	return simplejson.NewFromReader(r)
}

// Close closes the underlying websocket connection.
func (c *wsConn) Close() error {
	// unregister before finish.
	if c.user != nil {
		c.s.userBoard.Unregister(c.user)
	}
	return c.conn.Close()
}
