package ws

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"xim/broker"
	"xim/broker/proto"
	"xim/broker/userboard"
	"xim/broker/userds"

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
			fmt.Fprintln(w, "WS OK.")
		})
	}
	s.httpServer = &http.Server{
		Handler:      httpServeMux,
		Addr:         s.config.Addr,
		ReadTimeout:  s.config.HTTPReadTimeout,
		WriteTimeout: s.config.HTTPWriteTimeout,
	}
}

// ListenAndServe listens on the TCP network address s.confg.Addr
func (s *WebsocketServer) ListenAndServe() error {
	log.Println("http listening:", s.config.Addr)
	return s.httpServer.ListenAndServe()
}

func getAuthTokenFromRequest(r *http.Request) (token string, err error) {
	bearerAuth := r.Header.Get("Authorization")
	if bearerAuth != "" {
		parts := strings.SplitN(bearerAuth, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			return "", errors.New("invalid jwt authorization header=" + bearerAuth)
		}
		token = parts[1]
	} else {
		token = r.FormValue("jwt")
	}
	if token == "" {
		err = errors.New("need authorization token")
	}
	return
}

func (s *WebsocketServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	authToken, err := getAuthTokenFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	uid, err := userboard.VerifyAuthToken(authToken)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	instance := idPool.Get()
	defer idPool.Put(instance)

	user := &userds.UserLocation{
		UserIdentity: *uid,
		Broker:       s.config.Broker,
		Instance:     fmt.Sprintf("%d", instance),
	}

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	s.handleWebsocket(newWsConn(s, user, conn))
}

func (s *WebsocketServer) handleWebsocket(c *wsConn) {
	var (
		err          error
		logicMsgChan chan *proto.Msg
	)
	defer c.Close()

	log.Println("conn: ", c.conn.RemoteAddr())
	// authentication
	if err = registerUser(c); err != nil {
		log.Println(err)
		return
	}
	logicMsgChan = make(chan *proto.Msg, 10)
	go ProcessLogicMsg(c, logicMsgChan)
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
		if time.Now().Sub(t) > c.s.config.HeartbeatTimeout {
			log.Println("heartbeat timeout.")
			// ping timeout.
			return
		}

		log.Println(c.conn.RemoteAddr(), ":", msg.ID, msg.Type)
		switch msg.Type {
		case "":
			c.s.userBoard.Register(c.user, c)
			t = time.Now()
			c.WriteMsg(map[string]string{})
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
func ProcessLogicMsg(c *wsConn, q <-chan *proto.Msg) {
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

			replyMsg, err := broker.HandleLogicMsg(c.user, msg.Type, msg.Channel, msg.Kind, msg.Msg)
			// TODO handle send error.
			if err != nil {
				_ = c.WriteMsg(proto.NewErrorReply(msg.ID, err.Error()))
			} else if replyMsg != nil {
				_ = c.WriteMsg(proto.NewResponse(msg.ID, replyMsg))
			}
		}
	}
}

func registerUser(c *wsConn) (err error) {
	if err = c.s.userBoard.Register(c.user, c); err != nil {
		return err
	}
	if err = c.WriteMsg(proto.NewHello()); err != nil {
		return err
	}
	return
}
