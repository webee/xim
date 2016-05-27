package ws

import (
	"errors"
	"log"
	"time"

	"xim/broker/logic"
	"xim/broker/proto"
	"xim/broker/userboard"
	"xim/broker/userds"
	"xim/utils/msgutils"
)

// MsgLogic is handler of messages.
type MsgLogic struct {
	userBoard        *userboard.UserBoard
	user             *userds.UserLocation
	sender           msgutils.Sender
	closed           bool
	heartbeatTimeout time.Duration
}

// NewMsgLogic create a msg logic.
func NewMsgLogic(userBoard *userboard.UserBoard, user *userds.UserLocation, sender msgutils.Sender, heartbeatTimeout time.Duration) (*MsgLogic, error) {
	h := &MsgLogic{
		userBoard:        userBoard,
		user:             user,
		sender:           sender,
		heartbeatTimeout: heartbeatTimeout,
	}
	if err := h.register(); err != nil {
		log.Println(err)
		return nil, err
	}

	if err := h.PushMsg(&proto.Hello{User: user.User}); err != nil {
		log.Println(err)
		h.Close()
		return nil, err
	}
	return h, nil
}

// Handle handles the msg.
func (h *MsgLogic) Handle(msg msgutils.Message) bool {
	if h.closed {
		log.Println("client closed")
		return false
	}

	switch x := msg.(type) {
	case *proto.Null:
		h.register()
		h.PushMsg(x)
	case *proto.Ping:
		h.register()
		h.PushMsg(proto.PONG.New())
	case *proto.Bye:
		h.PushMsg(x)
		return false
	case *proto.Put:
		// handle by logic
		replyMsg, err := logic.PutMsg(h.user, x.Channel, x.Kind, x.Msg)
		// TODO handle send error.
		if err != nil {
			_ = h.PushMsg(proto.NewErrorReply(x.GetID(), err.Error()))
		} else if replyMsg != nil {
			_ = h.PushMsg(proto.NewReply(x.GetID(), replyMsg))
		}
	}
	return true
}

func (h *MsgLogic) register() error {
	return h.userBoard.Register(h.user, h)
}

// PushMsg push msg to msgbox.
func (h *MsgLogic) PushMsg(msg msgutils.Message) (err error) {
	if h.closed {
		return errors.New("client closed")
	}
	return h.sender.Send(msg)
}

// Close close this handler.
func (h *MsgLogic) Close() {
	// unregister before finish.
	if !h.closed {
		h.userBoard.Unregister(h.user)
		h.user.Close()
		h.closed = true
		log.Println(h.user, "msg handler closed.")
	}
}
