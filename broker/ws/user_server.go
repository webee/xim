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
	us.handleWebsocket(s, user, newWsConn(s, conn, 5))
}

func (us *UserServer) handleWebsocket(s *WebsocketServer, user *userds.UserLocation, c *wsConn) {
	go c.HandleMsg()
	defer c.Close()
	handler := NewMsgHandler(s.userBoard, user, c, s.config.HeartbeatTimeout, 5)
	handler.Start()
	defer handler.Close()

	for {
		msg, err := c.ReadMsg()
		if err != nil {
			log.Println(err)
			c.WriteMsg(proto.NewReply(nil, proto.ByeMsg, nil))
			break
		}

		if err := handler.HandleMsg(&msg); err != nil {
			break
		}
	}
}
