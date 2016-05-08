package ws

import (
	"fmt"
	"log"
	"net/http"

	"xim/broker/userboard"

	"github.com/gorilla/websocket"
)

// RequestHandler handles websocket request.
type RequestHandler interface {
	HandleRequest(s *WebsocketServer, w http.ResponseWriter, r *http.Request)
}

// WebsocketServer handles websocket connections.
type WebsocketServer struct {
	config         *WebsocketServerConfig
	upgrader       *websocket.Upgrader
	httpServer     *http.Server
	userBoard      *userboard.UserBoard
	requestHandler RequestHandler
}

// NewWebsocketServer creates a new WebsocketServer.
func NewWebsocketServer(userBoard *userboard.UserBoard, requestHandler RequestHandler, config *WebsocketServerConfig) (server *WebsocketServer) {
	server = &WebsocketServer{
		config:         config,
		userBoard:      userBoard,
		requestHandler: requestHandler,
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
	s.requestHandler.HandleRequest(s, w, r)
}
