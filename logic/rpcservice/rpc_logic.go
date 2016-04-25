package rpcservice

import (
	"errors"
	"log"

	"xim/broker/userboard"
	"xim/logic/dispatcher"
)

// RPCLogic represents the rpc logic.
type RPCLogic struct {
}

// RPCLogicHandleMsgArgs is the msg args.
type RPCLogicHandleMsgArgs struct {
	User    userboard.UserLocation
	Type    string
	Channel string
	Msg     interface{}
}

// RPCLogicHandleMsgReply is the msg reply.
type RPCLogicHandleMsgReply struct {
	Msg interface{}
}

// RPCServer methods.
const (
	PutMsgType        = "put"
	RPCLogicHandleMsg = "RPCLogic.HandleMsg"
)

// HandleMsg handle user send msg.
func (l *RPCLogic) HandleMsg(args *RPCLogicHandleMsgArgs, reply *RPCLogicHandleMsgReply) (err error) {
	log.Println(RPCLogicHandleMsg, "is called:", args.User, args.Type, args.Msg)
	switch args.Type {
	case PutMsgType:
		reply.Msg, err = handleMsgMsg(args.User, args.Channel, args.Msg)
	default:
		return errors.New(ErrUnknownMsgType)
	}
	return err
}

func handleMsgMsg(user userboard.UserLocation, channel string, msg interface{}) (replyMsg interface{}, err error) {
	// TODO
	// check org.user permission for channel.
	// errors.New(ErrPermDenied)
	if len(channel) < 3 {
		err = errors.New(ErrPermDenied)
		return
	}

	msgID, err := dispatcher.PutMsg(user, channel, msg)
	if err != nil {
		return
	}
	replyMsg = map[string]interface{}{
		"id": msgID,
	}
	return
}
