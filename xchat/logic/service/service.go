package service

import (
	"time"

	"xim/xchat/logic/db"

	"github.com/valyala/gorpc"
)

// services.
var (
	XChat = NewXChatService()
)

// SendMsgRequest is the send message method parameter.
type SendMsgRequest struct {
	ChatID uint64
	User   string
	Msg    string
}

// SendMsgReply is the send message method reply.
type SendMsgReply struct {
	MsgID uint64
	Ts    time.Time
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
func (s *XChatService) SendMsg(req *SendMsgRequest) (*SendMsgReply, error) {
	message, err := db.NewMsg(req.ChatID, req.User, req.Msg)
	if err != nil {
		return nil, err
	}

	// publish

	return &SendMsgReply{
		MsgID: message.MsgID,
		Ts:    message.Ts,
	}, nil
}
