package ws

import (
	"log"
	"net/http"
	"time"
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

	transeiver := msgutils.NewWSTranseiver(conn, new(proto.JSONObjSerializer), s.config.MsgBufSize)
	handler, err := NewMsgLogic(s.userBoard, user, transeiver, s.config.HeartbeatTimeout)
	if err != nil {
		return
	}
	defer handler.Close()

	t := time.After(s.config.HeartbeatTimeout)
	rc := transeiver.Receive()
	for {
		select {
		case msg, open := <-rc:
			if !open {
				return
			}

			if transeiver.Closed() {
				return
			}

			if ok := handler.Handle(msg); !ok {
				return
			}
		case <-t:
			if ok := handler.Handle(proto.PING.New()); !ok {
				return
			}
			t = time.After(s.config.HeartbeatTimeout)
		}
	}
}
