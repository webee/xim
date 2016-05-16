package ws

import (
	"errors"
	"log"

	"xim/broker"
	"xim/broker/proto"
	"xim/broker/userboard"
	"xim/broker/userds"
)

// MsgLogic is handler of messages.
type MsgLogic struct {
	userBoard *userboard.UserBoard
	user      *userds.UserLocation
	sender    Sender
	closed    bool
}

// NewMsgLogic create a msg logic.
func NewMsgLogic(userBoard *userboard.UserBoard, user *userds.UserLocation, sender Sender) (*MsgLogic, error) {
	h := &MsgLogic{
		userBoard: userBoard,
		user:      user,
		sender:    sender,
	}
	if err := h.register(); err != nil {
		log.Println(err)
		return nil, err
	}

	if err := h.PushMsg(proto.NewHello()); err != nil {
		log.Println(err)
		h.Close()
		return nil, err
	}
	return h, nil
}

// Handle handles the msg.
func (h *MsgLogic) Handle(msg *proto.Msg) bool {
	log.Println(h.user, ":", msg.SN, msg.Type, msg.Msg)
	if h.closed {
		log.Println("client closed")
		return false
	}

	switch msg.Type {
	case "":
		h.register()
		h.PushMsg(&proto.Reply{})
	case proto.PingMsg:
		// reseting user identity timeout.
		h.register()
		h.PushMsg(proto.NewPong(msg.Msg))
	case proto.ByeMsg:
		h.PushMsg(proto.NewBye())
		return false
	case proto.HelloMsg:
		// ignore
	default:
		// handle by logic
		replyMsg, err := broker.HandleLogicMsg(h.user, msg.Type, msg.Channel, msg.Kind, msg.Msg)
		// TODO handle send error.
		if err != nil {
			_ = h.PushMsg(proto.NewErrorReply(msg.SN, err.Error()))
		} else if replyMsg != nil {
			_ = h.PushMsg(proto.NewReply(msg.SN, replyMsg))
		}
	}
	return true
}

func (h *MsgLogic) register() error {
	return h.userBoard.Register(h.user, h)
}

// PushMsg push msg to msgbox.
func (h *MsgLogic) PushMsg(v interface{}) (err error) {
	if h.closed {
		return errors.New("client closed")
	}
	return h.sender.Send(v)
}

// Close close this handler.
func (h *MsgLogic) Close() {
	// unregister before finish.
	if h.user != nil {
		h.userBoard.Unregister(h.user)
		h.user.Close()
	}
	h.closed = true
	log.Println(h.user, "msg handler closed.")
}
