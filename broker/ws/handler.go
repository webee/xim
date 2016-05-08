package ws

import (
	"errors"
	"log"
	"time"

	"xim/broker"
	"xim/broker/proto"
	"xim/broker/userboard"
	"xim/broker/userds"
)

// UserConn is a user connection(ws/tcp).
type UserConn interface {
	PushMsg(user *userds.UserLocation, msg interface{}) error
}

// MsgHandler is handler of msges.
type MsgHandler struct {
	userBoard        *userboard.UserBoard
	user             *userds.UserLocation
	userConn         UserConn
	heartbeatTimeout time.Duration
	c                chan *proto.Msg
	msgbox           chan interface{}
	done             chan struct{}
}

// NewMsgHandler create a msg handler.
func NewMsgHandler(userBoard *userboard.UserBoard, user *userds.UserLocation, userConn UserConn, heartbeatTimeout time.Duration, chanSize int) *MsgHandler {
	return &MsgHandler{
		userBoard:        userBoard,
		user:             user,
		userConn:         userConn,
		heartbeatTimeout: heartbeatTimeout,
		c:                make(chan *proto.Msg, chanSize),
		msgbox:           make(chan interface{}, 5),
		done:             make(chan struct{}, 1),
	}
}

// Start starts this msg handler.
func (h *MsgHandler) Start() {
	go h.handle()
}

func (h *MsgHandler) handle() {
	var (
		err          error
		logicMsgChan chan *proto.Msg
	)

	logicMsgChan = make(chan *proto.Msg, 10)
	go h.ProcessLogicMsg(logicMsgChan)
	defer func() {
		close(logicMsgChan)
		<-h.done
	}()

	if err = h.Register(); err != nil {
		log.Println(err)
		return
	}

	if err = h.WriteMsg(proto.NewHello()); err != nil {
		log.Println(err)
		return
	}

	t := time.Now()
	for {
		msg, ok := <-h.c
		if !ok {
			break
		}

		if time.Now().Sub(t) > h.heartbeatTimeout {
			log.Println("heartbeat timeout.")
			// ping timeout.
			return
		}

		switch msg.Type {
		case "":
			h.Register()
			t = time.Now()
			h.PushMsg(map[string]string{})
		case proto.PingMsg:
			// reseting user identity timeout.
			h.Register()
			t = time.Now()
			h.PushMsg(proto.NewPong(msg.ID, msg.Msg))
		case proto.ByeMsg:
			h.PushMsg(proto.NewReplyBye(msg.ID))
			return
		case proto.HelloMsg:
			// ignore
		default:
			// handle by logic
			logicMsgChan <- msg
		}
	}
}

// ProcessLogicMsg process logic messages.
func (h *MsgHandler) ProcessLogicMsg(q <-chan *proto.Msg) {
	defer func() {
		close(h.done)
	}()

	for {
		select {
		case msg, ok := <-h.msgbox:
			// push
			if ok {
				h.WriteMsg(msg)
			}
		case msg, ok := <-q:
			// send
			if !ok {
				return
			}
			log.Println(h.user, ":", msg.ID, msg.Type, msg.Msg)

			replyMsg, err := broker.HandleLogicMsg(h.user, msg.Type, msg.Channel, msg.Kind, msg.Msg)
			// TODO handle send error.
			if err != nil {
				_ = h.WriteMsg(proto.NewErrorReply(msg.ID, err.Error()))
			} else if replyMsg != nil {
				_ = h.WriteMsg(proto.NewResponse(msg.ID, replyMsg))
			}
		}
	}
}

// Register register current user to userboard.
func (h *MsgHandler) Register() error {
	return h.userBoard.Register(h.user, h)
}

// HandleMsg put msg to channel.
func (h *MsgHandler) HandleMsg(msg *proto.Msg) (err error) {
	select {
	case <-h.done:
		return errors.New("handler closed")
	case h.c <- msg:
		return nil
	}
}

// WriteMsg push msg to connected user.
func (h *MsgHandler) WriteMsg(v interface{}) (err error) {
	return h.userConn.PushMsg(h.user, v)
}

// PushMsg push msg to msgbox.
func (h *MsgHandler) PushMsg(v interface{}) (err error) {
	h.msgbox <- v
	return nil
}

// Close close this handler.
func (h *MsgHandler) Close() {
	// unregister before finish.
	if h.user != nil {
		h.userBoard.Unregister(h.user)
	}
	close(h.c)
}
