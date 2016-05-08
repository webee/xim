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
		c:        newWsConn(s, conn, 100),
		app:      app,
		handlers: make(map[uint32]*MsgHandler),
	}
	ah.handleWebsocket()
}

// AppServerHandler handles app server.
type AppServerHandler struct {
	s        *WebsocketServer
	as       *AppServer
	c        *wsConn
	app      *userds.AppLocation
	handlers map[uint32]*MsgHandler
}

func (ah *AppServerHandler) handleWebsocket() {
	go ah.c.HandleMsg()
	defer ah.c.Close()
	defer ah.Close()

	ah.c.PushMsg(nil, proto.NewHello())
	for {
		msg, err := ah.c.ReadMsg()
		if err != nil {
			log.Println(err)
			ah.c.WriteMsg(proto.NewReply(nil, proto.ByeMsg, nil))
			break
		}

		switch msg.Type {
		case proto.AppRegisterUserMsg:
			user := ah.registerUser(msg.User)
			if user != nil {
				ah.c.PushMsg(nil, proto.NewReplyRegister(msg.ID, msg.User, user.Instance))
			}
		case proto.AppUnregisterUserMsg:
			if ah.unregisterUser(msg.UID) {
				ah.c.PushMsg(nil, proto.NewReplyUnregister(msg.ID, msg.UID))
			}
		case proto.ByeMsg:
			ah.c.PushMsg(nil, proto.NewReplyBye(msg.ID))
			return
		case proto.AppNullMsg:
		default:
			if msg.UID > 0 {
				if handler, ok := ah.handlers[msg.UID]; ok {
					handler.HandleMsg(&msg)
				}
			} else {
				for _, handler := range ah.handlers {
					handler.HandleMsg(&msg)
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
}

func (ah *AppServerHandler) registerUser(username string) *userds.UserLocation {
	if len(username) <= 0 {
		return nil
	}
	uid := &userds.UserIdentity{
		App:  ah.app.AppIdentity.App,
		User: username,
	}
	user := userds.NewUserLocation(uid, ah.s.config.Broker)

	handler := NewMsgHandler(ah.s.userBoard, user, ah, ah.s.config.HeartbeatTimeout, 5)
	ah.handlers[handler.user.Instance] = handler
	handler.Start()
	return user
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

// PushMsg write json message in a write timeout duration.
func (ah *AppServerHandler) PushMsg(user *userds.UserLocation, v interface{}) (err error) {
	handler, ok := ah.handlers[user.Instance]
	if !ok {
		return errors.New("user is disconnected")
	}
	switch msg := v.(type) {
	case *proto.ChannelMsg:
		msg.UID = handler.user.Instance
		return ah.c.PushMsg(nil, msg)
	case *proto.Reply:
		switch msg.Type {
		case proto.HelloMsg:
			// ignore hello.
			return nil
		}
		msg.UID = handler.user.Instance
		return ah.c.PushMsg(nil, msg)
	}
	return nil
}
