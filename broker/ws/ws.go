package ws

import (
	"fmt"
	"log"
	"net/http"

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
	go c.HandleMsg()
	defer c.Close()
	handler := NewMsgHandler(s.userBoard, c.user, c, s.config.HeartbeatTimeout, 5)
	handler.Start()
	defer handler.Close()

	for {
		msg, err := c.ReadMsg()
		if err != nil {
			log.Println(err)
			break
		}

		if err := handler.HandleMsg(&msg); err != nil {
			break
		}
	}
}
