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
	Name                   string
	MethodEcho             string
	MethodFetchChatMembers string
	MethodSendMsg          string
}

// NewXChatService creates a xchat service instance.
func NewXChatService() *XChatService {
	return &XChatService{
		Name:                   "XChat",
		MethodEcho:             "Echo",
		MethodFetchChatMembers: "FetchChatMembers",
		MethodSendMsg:          "SendMsg",
	}
}

// Echo send msg back.
func (s *XChatService) Echo(msg string) string {
	return msg
}

// FetchChatMembers fetch chat's members.
func (s *XChatService) FetchChatMembers(chatID uint64) ([]db.Member, error) {
	return db.GetChatMembers(chatID)
}

// SendMsg sends message.
func (s *XChatService) SendMsg(req *SendMsgRequest) (*db.Message, error) {
	message, err := db.NewMsg(req.ChatID, req.User, req.Msg)
	if err != nil {
		return nil, err
	}

	return message, nil
}
