package ws

import (
	"errors"
	"log"
	"net/http"
	"xim/broker/proto"
	"xim/broker/userboard"
	"xim/broker/userds"
)

// AppServer handles user websocket connection.
type AppServer struct {
}

// NewAppServer create a app server.
func NewAppServer() *AppServer {
	return &AppServer{}
}

// HandleRequest handles the websocket request.
func (as *AppServer) HandleRequest(s *WebsocketServer, w http.ResponseWriter, r *http.Request) {
	authToken, err := getAuthTokenFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	aid, err := userboard.VerifyAppAuthToken(authToken)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	app := userds.NewAppLocation(aid, s.config.Broker)
	defer app.Close()

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ah := &AppServerHandler{
		s:        s,
		as:       as,
		c:        newWsConnection(conn, 100, s.config.HeartbeatTimeout, s.config.WriteTimeout),
		app:      app,
		handlers: make(map[uint32]*MsgLogic),
	}
	ah.handleWebsocket()
}

// AppServerHandler handles app server.
type AppServerHandler struct {
	s        *WebsocketServer
	as       *AppServer
	c        *wsConnection
	app      *userds.AppLocation
	handlers map[uint32]*MsgLogic
}

func (ah *AppServerHandler) handleWebsocket() {
	defer ah.Close()
	c := ah.c
	c.Send(proto.NewHello())
	r := c.Receive()

	for {
		var msg *proto.Msg
		var open bool
		msg, open = <-r
		if !open {
			return
		}

		switch msg.Type {
		case proto.AppRegisterUserMsg:
			user, err := ah.registerUser(msg.User)
			if err != nil {
				log.Println("register user error:", err)
				continue
			}
			if user != nil {
				c.Send(proto.NewReplyRegister(msg.SN, msg.User, user.Instance))
			}
		case proto.AppUnregisterUserMsg:
			if ah.unregisterUser(msg.UID) {
				c.Send(proto.NewReplyUnregister(msg.SN, msg.UID))
			}
		case proto.ByeMsg:
			c.Send(proto.NewBye())
			return
		case proto.AppNullMsg:
		default:
			if msg.UID > 0 {
				if handler, ok := ah.handlers[msg.UID]; ok {
					_ = handler.Handle(msg)
				}
			} else {
				for _, handler := range ah.handlers {
					_ = handler.Handle(msg)
				}
			}
		}
	}
}

// Close close all resources.
func (ah *AppServerHandler) Close() {
	for _, handler := range ah.handlers {
		handler.Close()
	}
	ah.c.Close()
}

func (ah *AppServerHandler) registerUser(username string) (*userds.UserLocation, error) {
	if len(username) <= 0 {
		return nil, errors.New("bad username")
	}
	uid := &userds.UserIdentity{
		App:  ah.app.AppIdentity.App,
		User: username,
	}
	user := userds.NewUserLocation(uid, ah.s.config.Broker)

	handler, err := NewMsgLogic(ah.s.userBoard, user, newAppUserSender(ah, user, ah.c), ah.s.config.HeartbeatTimeout)
	if err != nil {
		return nil, err
	}
	ah.handlers[handler.user.Instance] = handler
	return user, nil
}

func (ah *AppServerHandler) unregisterUser(uid uint32) bool {
	handler := ah.handlers[uid]
	if handler != nil {
		delete(ah.handlers, uid)
		handler.Close()
		return true
	}
	return false
}

// AppUserSender can send msg to user.
type AppUserSender struct {
	app    *AppServerHandler
	user   *userds.UserLocation
	sender Sender
}

func newAppUserSender(app *AppServerHandler, user *userds.UserLocation, sender Sender) Sender {
	return &AppUserSender{
		app:    app,
		user:   user,
		sender: sender,
	}
}

// Send sends msg to user.
func (s *AppUserSender) Send(v interface{}) error {
	switch msg := v.(type) {
	case *proto.ChannelMsg:
		msg.UID = s.user.Instance
		return s.sender.Send(msg)
	case *proto.TypeMsg:
		switch msg.Type {
		case proto.HelloMsg:
			// ignore hello.
			return nil
		case proto.ByeMsg:
			s.app.unregisterUser(s.user.Instance)
		}
		msg.UID = s.user.Instance
		return s.sender.Send(msg)
	case *proto.Reply:
		switch msg.Type {
		case proto.HelloMsg:
			// ignore hello.
			return nil
		}
		msg.UID = s.user.Instance
		return s.sender.Send(msg)
	}
	return nil
}
