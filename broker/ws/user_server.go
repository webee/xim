package ws

import (
	"log"
	"net/http"
	"xim/broker/proto"
	"xim/broker/userboard"
	"xim/broker/userds"
	"xim/utils/msgutils"
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
	transeiver := msgutils.NewWSTranseiver(conn, new(proto.JSONObjSerializer), s.config.MsgBufSize, s.config.HeartbeatTimeout)
	defer transeiver.Close()

	us.handleWebsocket(s, user, transeiver)
}

func (us *UserServer) handleWebsocket(s *WebsocketServer, user *userds.UserLocation, transeiver msgutils.Transeiver) {
	handler, err := NewMsgLogic(s.userBoard, user, transeiver, s.config.HeartbeatTimeout)
	if err != nil {
		return
	}
	defer handler.Close()

	r := transeiver.Receive()

	var msg msgutils.Message
	//var closed bool
	var open bool
	for {
		msg, open = <-r
		if !open {
			return
		}

		/*
			if !closed && transeiver.Closed() {
				handler.unregister()
			}
		*/
		if ok := handler.Handle(msg); !ok {
			break
		}
	}
}
