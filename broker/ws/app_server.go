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
	transeiver := msgutils.NewWSTranseiver(conn, new(proto.JSONObjSerializer), 1000*s.config.MsgBufSize, s.config.HeartbeatTimeout, proto.PONG.New())
	defer transeiver.Close()

	ah := &AppServerHandler{
		s:          s,
		as:         as,
		transeiver: transeiver,
		app:        app,
		handlers:   make(map[uint32]*MsgLogic),
	}
	defer ah.Close()

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
	transeiver := ah.transeiver
	if err := transeiver.Send(&proto.Hello{App: ah.app.App}); err != nil {
		return
	}

	rc := transeiver.Receive()
	for {
		msg, open := <-rc
		if !open {
			return
		}

		switch x := msg.(type) {
		case *proto.Bye:
			return
		case *proto.Register:
			user, err := ah.registerUser(x.User)
			works <- func() {
				if err != nil {
					log.Println("register user error:", err)
					transeiver.Send(proto.NewErrorReply(x.GetID(), "register error"))
					return
				}
				if user == nil {
					transeiver.Send(proto.NewErrorReply(x.GetID(), "register failed"))
					return
				}
				transeiver.Send(proto.NewAppReply(x.GetID(), user.Instance, nil))
			}
		case *proto.Unregister:
			ok := ah.unregisterUser(x.UID)
			works <- func() {
				if ok {
					transeiver.Send(proto.NewAppReply(x.GetID(), x.UID, nil))
				} else {
					transeiver.Send(proto.NewAppErrorReply(x.GetID(), x.UID, "unregister failed"))
				}
			}
		case *proto.Null:
			handlers := ah.getHandlers(0)
			works <- func() {
				for _, handler := range handlers {
					handler.Handle(x)
				}
				transeiver.Send(x)
			}
		case *proto.Ping:
			handlers := ah.getHandlers(0)
			works <- func() {
				for _, handler := range handlers {
					handler.Handle(x)
				}
				transeiver.Send(proto.PONG.New())
			}
		case *proto.Pong:
			handlers := ah.getHandlers(0)
			works <- func() {
				for _, handler := range handlers {
					handler.Handle(x)
				}
			}
		case *proto.Put:
			handlers := ah.getHandlers(x.UID)
			works <- func() {
				for _, handler := range handlers {
					handler.Handle(x)
				}
			}
		default:
			log.Println("bad message type:", msg.MessageType())
		}
	}
}

func (ah *AppServerHandler) getHandlers(uid uint32) []*MsgLogic {
	handlers := []*MsgLogic{}
	if uid > 0 {
		if handler, ok := ah.handlers[uid]; ok {
			handlers = append(handlers, handler)
		}
	} else {
		for _, handler := range ah.handlers {
			handlers = append(handlers, handler)
		}
	}
	return handlers
}

// Close close all resources.
func (ah *AppServerHandler) Close() {
	for _, handler := range ah.handlers {
		handler.Close()
	}
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

	handler, err := NewMsgLogic(ah.s.userBoard, user, newAppUserSender(ah, user, ah.transeiver))
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
func (s *AppUserSender) Send(msg msgutils.Message) error {
	switch x := msg.(type) {
	case *proto.Push:
		x.UID = s.user.Instance
		return s.sender.Send(x)
	case *proto.Hello:
	case *proto.Bye:
		// impossible.
		//s.app.unregisterUser(s.user.Instance)
	case *proto.Reply:
		x.UID = s.user.Instance
		return s.sender.Send(x)
	}
	return nil
}
