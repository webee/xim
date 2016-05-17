package ws

import (
	"errors"
	"log"
	"net/http"
	"xim/broker/proto"
	"xim/broker/userboard"
	"xim/broker/userds"
	"xim/utils/msgutils"
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
		s:          s,
		as:         as,
		transeiver: msgutils.NewWSTranseiver(conn, new(ProtoJSONSerializer), 100*s.config.MsgBufSize, s.config.HeartbeatTimeout),
		app:        app,
		handlers:   make(map[uint32]*MsgLogic),
	}
	log.Printf("app: %s connected.", app)
	defer func() {
		log.Printf("app: %s disconnected.", app)
	}()

	ah.handleWebsocket()
}

// AppServerHandler handles app server.
type AppServerHandler struct {
	s          *WebsocketServer
	as         *AppServer
	transeiver msgutils.Transeiver
	app        *userds.AppLocation
	handlers   map[uint32]*MsgLogic
}

func (ah *AppServerHandler) handleWebsocket() {
	defer ah.Close()
	transeiver := ah.transeiver
	transeiver.Send(proto.NewHello())
	r := transeiver.Receive()

	for {
		var msg *proto.Msg
		var m msgutils.Message
		var open bool
		m, open = <-r
		if !open {
			return
		}
		msg = m.(*proto.Msg)

		switch msg.Type {
		case proto.AppRegisterUserMsg:
			user, err := ah.registerUser(msg.User)
			if err != nil {
				log.Println("register user error:", err)
				transeiver.Send(proto.NewErrorReply(msg.SN, "register error"))
				continue
			}
			if user == nil {
				transeiver.Send(proto.NewErrorReply(msg.SN, "register failed"))
				continue
			}
			transeiver.Send(proto.NewAppReply(user.Instance, msg.SN, nil))
		case proto.AppUnregisterUserMsg:
			if ah.unregisterUser(msg.UID) {
				transeiver.Send(proto.NewAppReply(msg.UID, msg.SN, nil))
			} else {
				transeiver.Send(proto.NewAppErrorReply(msg.UID, msg.SN, "unregister failed"))
			}
		case proto.ByeMsg:
			transeiver.Send(proto.NewBye())
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
	ah.transeiver.Close()
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

	handler, err := NewMsgLogic(ah.s.userBoard, user,
		newAppUserSender(ah, user, ah.transeiver),
		ah.s.config.HeartbeatTimeout)
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
	sender msgutils.Sender
}

func newAppUserSender(app *AppServerHandler, user *userds.UserLocation, sender msgutils.Sender) msgutils.Sender {
	return &AppUserSender{
		app:    app,
		user:   user,
		sender: sender,
	}
}

// Send sends msg to user.
func (s *AppUserSender) Send(v msgutils.Message) error {
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
