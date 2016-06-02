package service

import (
	"xim/xchat/logic/db"
	"xim/xchat/logic/logger"

	ol "github.com/go-ozzo/ozzo-log"

	"github.com/valyala/gorpc"
)

// variables
var (
	l     *ol.Logger
	XChat = NewXChatService()
)

// Init setup router.
func Init() {
	l = logger.Logger.GetLogger("service")
}

// SendMsgRequest is the send message method parameter.
type SendMsgRequest struct {
	ChatID uint64
	User   string
	Msg    string
}

// NewServiceDispatcher creates the services dispatcher.
func NewServiceDispatcher() *gorpc.Dispatcher {
	d := gorpc.NewDispatcher()
	d.AddService(XChat.Name, XChat)
	return d
}

// XChatService provide xchat services.
type XChatService struct {
	Name          string
	MethodEcho    string
	MethodSendMsg string
}

// NewXChatService creates a xchat service instance.
func NewXChatService() *XChatService {
	return &XChatService{
		Name:          "XChat",
		MethodEcho:    "Echo",
		MethodSendMsg: "SendMsg",
	}
}

// Echo send msg back.
func (s *XChatService) Echo(msg string) string {
	return msg
}

// SendMsg sends message.
func (s *XChatService) SendMsg(clientAddr string, req *SendMsgRequest) (*db.Message, error) {
	l.Debug("%s send message", clientAddr)
	message, err := db.NewMsg(req.ChatID, req.User, req.Msg)
	if err != nil {
		return nil, err
	}

	// publish

	return message, nil
}
