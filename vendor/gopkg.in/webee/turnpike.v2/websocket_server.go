package turnpike

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	jsonWebsocketProtocol    = "wamp.2.json"
	msgpackWebsocketProtocol = "wamp.2.msgpack"
)

type invalidPayload byte

func (e invalidPayload) Error() string {
	return fmt.Sprintf("Invalid payloadType: %d", e)
}

type protocolExists string

func (e protocolExists) Error() string {
	return "This protocol has already been registered: " + string(e)
}

type protocol struct {
	payloadType int
	serializer  Serializer
}

// ConnectionConfig is the configs of a connection.
type ConnectionConfig struct {
	MaxMsgSize   int64
	WriteTimeout time.Duration
	PingTimeout  time.Duration
	IdleTimeout  time.Duration
}

// WebsocketServer handles websocket connections.
type WebsocketServer struct {
	Router
	Upgrader *websocket.Upgrader

	protocols map[string]protocol

	// The serializer to use for text frames. Defaults to JSONSerializer.
	TextSerializer Serializer
	// The serializer to use for binary frames. Defaults to JSONSerializer.
	BinarySerializer Serializer
	ConnectionConfig
}

// NewWebsocketServer creates a new WebsocketServer from a map of realms
func NewWebsocketServer(realms map[string]*Realm) (*WebsocketServer, error) {
	tlog.Println("NewWebsocketServer")
	r := NewDefaultRouter()
	for uri, realm := range realms {
		if err := r.RegisterRealm(URI(uri), realm); err != nil {
			return nil, err
		}
	}
	s := newWebsocketServer(r)
	return s, nil
}

// NewBasicWebsocketServer creates a new WebsocketServer with a single basic realm
func NewBasicWebsocketServer(uri string) *WebsocketServer {
	tlog.Println("NewBasicWebsocketServer")
	s, _ := NewWebsocketServer(map[string]*Realm{uri: {}})
	return s
}

func newWebsocketServer(r Router) *WebsocketServer {
	s := &WebsocketServer{
		Router:    r,
		protocols: make(map[string]protocol),
		ConnectionConfig: ConnectionConfig{
			PingTimeout: 3 * time.Minute,
		},
	}
	s.Upgrader = &websocket.Upgrader{}
	s.RegisterProtocol(jsonWebsocketProtocol, websocket.TextMessage, new(JSONSerializer))
	s.RegisterProtocol(msgpackWebsocketProtocol, websocket.BinaryMessage, new(MessagePackSerializer))
	return s
}

// RegisterProtocol registers a serializer that should be used for a given protocol string and payload type.
func (s *WebsocketServer) RegisterProtocol(proto string, payloadType int, serializer Serializer) error {
	tlog.Println("RegisterProtocol:", proto)
	if payloadType != websocket.TextMessage && payloadType != websocket.BinaryMessage {
		return invalidPayload(payloadType)
	}
	if _, ok := s.protocols[proto]; ok {
		return protocolExists(proto)
	}
	s.protocols[proto] = protocol{payloadType, serializer}
	s.Upgrader.Subprotocols = append(s.Upgrader.Subprotocols, proto)
	return nil
}

func (s *WebsocketServer) GetLocalClientWithSize(sz int, realm string, details map[string]interface{}) (*Client, error) {
	peer, err := s.Router.GetLocalPeerWithSize(sz, URI(realm), details)
	if err != nil {
		return nil, err
	}
	c := NewClient(peer)
	go c.Receive()
	return c, nil
}

// GetLocalClient returns a client connected to the specified realm
func (s *WebsocketServer) GetLocalClient(realm string, details map[string]interface{}) (*Client, error) {
	return s.GetLocalClientWithSize(100, realm, details)
}

// ServeHTTP handles a new HTTP connection.
func (s *WebsocketServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tlog.Println("WebsocketServer.ServeHTTP", r.Method, r.RequestURI)
	// TODO: subprotocol?
	conn, err := s.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		tlog.Println("Error upgrading to websocket connection:", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	s.handleWebsocket(conn)
}

func (s *WebsocketServer) handleWebsocket(conn *websocket.Conn) {
	var serializer Serializer
	var payloadType int
	if proto, ok := s.protocols[conn.Subprotocol()]; ok {
		serializer = proto.serializer
		payloadType = proto.payloadType
	} else {
		// TODO: this will not currently ever be hit because
		//       gorilla/websocket will reject the conncetion
		//       if the subprotocol isn't registered
		switch conn.Subprotocol() {
		case jsonWebsocketProtocol:
			serializer = new(JSONSerializer)
			payloadType = websocket.TextMessage
		case msgpackWebsocketProtocol:
			serializer = new(MessagePackSerializer)
			payloadType = websocket.BinaryMessage
		default:
			conn.Close()
			return
		}
	}

	peer := websocketPeer{
		conn:             conn,
		serializer:       serializer,
		sendMsgs:         make(chan Message, 16),
		messages:         make(chan Message, 10),
		payloadType:      payloadType,
		closing:          make(chan struct{}),
		ConnectionConfig: &s.ConnectionConfig,
	}
	go peer.run()

	logErr(s.Router.Accept(&peer))
}
