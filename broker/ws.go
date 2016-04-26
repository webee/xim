package broker

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"xim/broker/proto"
	"xim/broker/userboard"

	"github.com/bitly/go-simplejson"
	"github.com/gorilla/websocket"
	"github.com/youtube/vitess/go/pools"
)

var (
	idPool = pools.NewIDPool()
)

// WebsocketServer handles websocket connections.
type WebsocketServer struct {
	config     *WebsocketServerConfig
	upgrader   *websocket.Upgrader
	httpServer *http.Server
	userBoard  *userboard.UserBoard
}

// WsConn is a websocket connection.
type WsConn struct {
	wLock    sync.Mutex
	s        *WebsocketServer
	conn     *websocket.Conn
	user     *userboard.UserLocation
	from     string
	instance uint32
	msgbox   chan interface{}
	done     chan struct{}
}

// NewWebsocketServer creates a new WebsocketServer.
func NewWebsocketServer(userBoard *userboard.UserBoard, config *WebsocketServerConfig) (server *WebsocketServer) {
	server = &WebsocketServer{
		config:    config,
		userBoard: userBoard,
	}
	server.initUpgrader()
	server.initHTTPServer()
	return server
}

// NewWsConn creates a WsConn.
func NewWsConn(s *WebsocketServer, conn *websocket.Conn) *WsConn {
	return &WsConn{
		s:        s,
		conn:     conn,
		from:     conn.RemoteAddr().String(),
		instance: idPool.Get(),
		msgbox:   make(chan interface{}, 5),
		done:     make(chan struct{}, 1),
	}
}

func (s *WebsocketServer) initUpgrader() {
	s.upgrader = &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
}

func (s *WebsocketServer) initHTTPServer() {
	httpServeMux := http.NewServeMux()
	httpServeMux.Handle("/ws", s)
	if s.config.Testing {
		httpServeMux.HandleFunc("/testing", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "OK")
		})
	}
	s.httpServer = &http.Server{
		Handler:      httpServeMux,
		Addr:         s.config.Addr,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
}

// ListenAndServe listens on the TCP network address s.confg.Addr
func (s *WebsocketServer) ListenAndServe() error {
	log.Println("http listening:", s.config.Addr)
	return s.httpServer.ListenAndServe()
}

func (s *WebsocketServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	s.handleWebsocket(NewWsConn(s, conn))
}

func (s *WebsocketServer) handleWebsocket(c *WsConn) {
	var (
		err          error
		logicMsgChan chan *proto.Msg
	)
	defer c.Close()

	log.Println("conn: ", c.conn.RemoteAddr())
	// authentication
	if err = c.authenticate(); err != nil {
		log.Println(err)
		return
	}
	logicMsgChan = make(chan *proto.Msg, 10)
	go c.ProcessLogicMsg(logicMsgChan)
	defer func() {
		close(logicMsgChan)
		<-c.done
	}()

	t := time.Now()
	for {
		// read msgs.
		msg, err := c.ReadMsg()
		if err != nil {
			log.Println(err)
			break
		}
		if time.Now().Sub(t) > c.s.config.HeartBeatTimeout {
			// ping timeout.
			return
		}

		log.Println(c.conn.RemoteAddr(), ":", msg.ID, msg.Type)
		switch msg.Type {
		case proto.PingMsg:
			// reseting user identity timeout.
			c.s.userBoard.Register(c.user, c)
			t = time.Now()
			c.WriteMsg(proto.NewPong(msg.ID, msg.Msg))
		case proto.ByeMsg:
			c.WriteMsg(proto.NewReplyBye(msg.ID))
			return
		case proto.HelloMsg:
			// ignore
		default:
			// handle by logic
			logicMsgChan <- &msg
		}
	}
}

// ProcessLogicMsg process logic messages.
func (c *WsConn) ProcessLogicMsg(q <-chan *proto.Msg) {
	defer func() {
		log.Println(c.conn.RemoteAddr(), ":", "OVER.")
		close(c.done)
	}()

	for {
		select {
		case msg, ok := <-c.msgbox:
			// push
			if ok {
				c.WriteMsg(msg)
			}
		case msg, ok := <-q:
			// send
			if !ok {
				return
			}
			log.Println(c.conn.RemoteAddr(), ":", msg.ID, msg.Type, msg.Msg)

			replyMsg, err := HandleLogicMsg(c.user, msg.Type, msg.Channel, msg.Kind, msg.Msg)
			// TODO handle send error.
			if err != nil {
				_ = c.WriteMsg(proto.NewErrorReply(msg.ID, err.Error()))
			} else if replyMsg != nil {
				_ = c.WriteMsg(proto.NewResponse(msg.ID, replyMsg))
			}
		}
	}
}

func (c *WsConn) authenticate() (err error) {
	msg := proto.Msg{}
	if _, err = c.ReadJSON(&msg, c.s.config.AuthTimeout); err != nil {
		return
	}
	if msg.Type != proto.HelloMsg {
		return errors.New("need hello msg")
	}

	token := msg.Token
	log.Println("token: ", token)
	if uid, err := userboard.VerifyAuthToken(token); err == nil {
		log.Println(c.from, "auth ok.")
		c.user = &userboard.UserLocation{
			UserIdentity: *uid,
			Broker:       c.s.config.Broker,
			Instance:     fmt.Sprintf("%d", c.instance),
		}
		if err = c.s.userBoard.Register(c.user, c); err != nil {
			return err
		}
		if err = c.WriteMsg(proto.NewWelcome(msg.ID)); err != nil {
			return err
		}
	}
	return
}

// ReadMsg read json message in a heartbeat duration.
func (c *WsConn) ReadMsg() (msg proto.Msg, err error) {
	//jd, err = c.ReadJSONData(c.s.config.HeartBeatTimeout)
	//msg.ID, err = jd.Get("id").Int()
	//msg.Type, err = jd.Get("type").String()

	// msg with bytes
	/*
		msg = new(proto.MsgWithBytes)
		bytes, err := c.ReadJSON(msg, c.s.config.HeartBeatTimeout)
		msg.Bytes = bytes
	*/
	_, err = c.ReadJSON(&msg, c.s.config.HeartBeatTimeout)
	return
}

// PushMsg write json message in a write timeout duration.
func (c *WsConn) PushMsg(v interface{}) (err error) {
	select {
	case <-c.done:
		return errors.New("connection closed")
	case c.msgbox <- v:
		return nil
	}
}

// WriteMsg write json message in a write timeout duration.
func (c *WsConn) WriteMsg(v interface{}) (err error) {
	err = c.WriteJSON(v, c.s.config.WriteTimeout)
	return
}

// WriteJSON write json message in a timeout duration.
func (c *WsConn) WriteJSON(v interface{}, timeout time.Duration) error {
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
func (c *WsConn) ReadJSON(v interface{}, timeout time.Duration) (bytes []byte, err error) {
	conn := c.conn
	conn.SetReadDeadline(time.Now().Add(timeout))
	_, bytes, err = conn.ReadMessage()
	if len(bytes) > 16*1024 {
		err = errors.New("msg too large")
		return
	}
	conn.SetReadDeadline(time.Time{})
	if err != nil {
		return
	}
	err = json.Unmarshal(bytes, v)
	return
}

// ReadJSONData read json message in a timeout duration.
func (c *WsConn) ReadJSONData(timeout time.Duration) (jd *simplejson.Json, err error) {
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
func (c *WsConn) Close() error {
	// unregister before finish.
	if c.user != nil {
		c.s.userBoard.Unregister(c.user)
	}
	idPool.Put(c.instance)
	return c.conn.Close()
}
