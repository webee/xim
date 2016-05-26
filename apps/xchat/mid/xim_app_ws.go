package mid

import (
	"errors"
	"fmt"
	"net/http"
	"time"
	"xim/broker/proto"
	"xim/utils/msgutils"

	"github.com/gorilla/websocket"
)

func getWSTranseiver(url, token string, msgBufSize int) (msgutils.Transeiver, error) {
	conn, err := getWSConn(url, token)
	if err != nil {
		return nil, err
	}
	transeiver := msgutils.NewWSTranseiver(conn, new(proto.JSONObjSerializer), msgBufSize, 0)

	// waiting hello
	msg, err := msgutils.GetMessageTimeout(transeiver, 2*time.Second)
	if err != nil {
		return nil, fmt.Errorf("waiting hello msg error: %s", err)
	}
	if _, ok := msg.(*proto.Hello); !ok {
		return nil, errors.New("waiting hello msg failed")
	}

	return transeiver, nil
}

func getWSConn(url, token string) (*websocket.Conn, error) {
	header := http.Header{}
	header.Add("Authorization", "Bearer "+token)
	conn, _, err := websocket.DefaultDialer.Dial(url, header)
	return conn, err
}

// XIMAppWsController is a app websocket connection controller.
type XIMAppWsController struct {
	*msgutils.MsgController
}

// NewXIMAppWsController creates a xim app websocket connection controller.
func NewXIMAppWsController(t msgutils.Transeiver, handler msgutils.MessageHandler, closeHandler msgutils.CloseHandler) *XIMAppWsController {
	return &XIMAppWsController{
		MsgController: msgutils.NewMsgController(t, handler, closeHandler),
	}
}

// Send send a msg.
func (x *XIMAppWsController) Send(msg msgutils.Message) error {
	return x.MsgController.Send(msg)
}

// Req send a msg and wait reply.
func (x *XIMAppWsController) Req(msg msgutils.SyncMessage) (msgutils.SyncMessage, error) {
	return x.SyncSend(msg)
}

// Close close the controller.
func (x *XIMAppWsController) Close() error {
	return x.MsgController.Close()
}
