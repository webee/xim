package ws

import (
	"log"
	"net/http"
	"xim/broker/proto"
	"xim/broker/userboard"
	"xim/broker/userds"
)

// UserServer handles user websocket connection.
type UserServer struct {
}

// HandleRequest handles the websocket request.
func (us *UserServer) HandleRequest(s *WebsocketServer, w http.ResponseWriter, r *http.Request) {
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
	user := userds.NewUserLocation(uid, s.config.Broker)
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	c := newWsConnection(conn, 5, s.config.HeartbeatTimeout, s.config.WriteTimeout)
	us.handleWebsocket(s, user, c)
}

func (us *UserServer) handleWebsocket(s *WebsocketServer, user *userds.UserLocation, c Connection) {
	defer c.Close()
	handler, err := NewMsgLogic(s.userBoard, user, c, s.config.HeartbeatTimeout)
	if err != nil {
		return
	}
	defer handler.Close()

	r := c.Receive()
	for {
		var msg *proto.Msg
		var open bool
		msg, open = <-r
		if !open {
			return
		}
		if ok := handler.Handle(msg); !ok {
			break
		}
	}
}
