package broker

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"xim/broker/proto"

	"github.com/bitly/go-simplejson"
	"github.com/gorilla/websocket"
)

// WebsocketServer handles websocket connections.
type WebsocketServer struct {
	config     *WebsocketServerConfig
	upgrader   *websocket.Upgrader
	httpServer *http.Server
	userBoard  *UserBoard
}

// WsConn is a websocket connection.
type WsConn struct {
	wLock sync.Mutex
	s     *WebsocketServer
	conn  *websocket.Conn
	uid   *UserIdentity
	from  string
}

// NewWebsocketServer creates a new WebsocketServer.
func NewWebsocketServer(userBoard *UserBoard, config *WebsocketServerConfig) (server *WebsocketServer) {
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
		s:    s,
		conn: conn,
		from: conn.RemoteAddr().String(),
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
		err           error
		brokerMsgChan = make(chan *proto.MsgWithBytes, 15)
		logicMsgChan  = make(chan *proto.MsgWithBytes, 10)
		logicFinish   = make(chan bool, 1)
	)

	log.Println("conn: ", c.conn.RemoteAddr())
	// authentication
	if err = c.authenticate(); err != nil {
		log.Println(err)
		return
	}
	// unregister before finish.
	defer c.s.userBoard.Unregister(c.uid, c.from)

	// read msgs.
	go func() {
		for {
			msg, err := c.ReadMsg()
			if err != nil {
				log.Println(err)
				brokerMsgChan <- nil
				return
			}
			log.Println(c.conn.RemoteAddr(), ":", msg.ID, msg.Type)
			brokerMsgChan <- msg
		}
	}()
	go c.ProcessLogicMsg(logicMsgChan, logicFinish)
	defer func() {
		_ = <-logicFinish
		c.Close()
	}()
	defer func() {
		logicMsgChan <- nil
	}()

	for {
		msg := <-brokerMsgChan
		if msg == nil {
			return
		}
		switch msg.Type {
		case proto.PingMsg:
			// reseting user identity timeout.
			c.WriteMsg(proto.NewPong(msg.ID))
		case proto.ByeMsg:
			c.WriteMsg(proto.NewReplyBye(msg.ID))
			return
		case proto.HelloMsg:
			// ignore
		default:
			// handle by logic
			logicMsgChan <- msg
		}
	}
}

// ProcessLogicMsg process logic messages.
func (c *WsConn) ProcessLogicMsg(q <-chan *proto.MsgWithBytes, finish chan bool) {
	for {
		msg := <-q
		if msg == nil {
			log.Println(c.conn.RemoteAddr(), ":", "OVER.")
			finish <- true
			return
		}

		log.Println(c.conn.RemoteAddr(), ":", msg.ID, msg.Type, string(msg.Bytes))

		bytes, err := HandleLogicMsg("#1", "AAA", "webee", "#1", msg.Bytes)
		// TODO handle send error.
		if err != nil {
			_ = c.conn.WriteJSON(proto.NewReplyError(msg.ID, msg.Type, "error"))
		} else {
			_ = c.conn.WriteMessage(websocket.TextMessage, bytes)
		}
	}
}

func (c *WsConn) authenticate() (err error) {
	hello, err := c.ReadHello()
	if err != nil {
		return
	}
	log.Println("token: ", hello.Token)
	c.uid, err = VerifyAuthToken(hello.Token)
	if err == nil {
		log.Println(c.from, "auth ok.")
		err = c.s.userBoard.Register(c.uid, c.from, c)
	}
	err = c.WriteMsg(proto.NewWelcome(hello.ID))

	return
}

// ReadMsg read json message in a heartbeat duration.
func (c *WsConn) ReadMsg() (msg *proto.MsgWithBytes, err error) {
	//jd, err = c.ReadJSONData(c.s.config.HeartBeatTimeout)
	//msg.ID, err = jd.Get("id").Int()
	//msg.Type, err = jd.Get("type").String()
	msg = new(proto.MsgWithBytes)
	bytes, err := c.ReadJSON(msg, c.s.config.HeartBeatTimeout)
	msg.Bytes = bytes
	return
}

// ReadHello read the hello message in a auth timeout duration.
func (c *WsConn) ReadHello() (hello proto.Hello, err error) {
	_, err = c.ReadJSON(&hello, c.s.config.AuthTimeout)
	return
}

// WriteMsg write json message in a write timeout duration.
func (c *WsConn) WriteMsg(v interface{}) (err error) {
	err = c.WriteJSON(v, c.s.config.WriteTimeout)
	return
}

// WriteJSON write json message in a timeout duration.
func (c *WsConn) WriteJSON(v interface{}, timeout time.Duration) error {
	conn := c.conn
	c.wLock.Lock()
	conn.SetWriteDeadline(time.Now().Add(timeout))
	err := conn.WriteJSON(v)
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
	return c.conn.Close()
}
