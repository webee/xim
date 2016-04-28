package rpcservice

import (
	"errors"
	"log"

	"xim/broker/proto"
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
	Kind    string
	Msg     interface{}
}

// RPCLogicHandleMsgReply is the msg reply.
type RPCLogicHandleMsgReply struct {
	Msg interface{}
}

// RPCLogicVerifyUserTokenArgs is the msg args.
type RPCLogicVerifyUserTokenArgs struct {
	App   string
	Token string
}

// RPCLogicVerifyUserTokenReply is the msg reply.
type RPCLogicVerifyUserTokenReply struct {
	uid userboard.UserIdentity
}

// RPCServer methods.
const (
	RPCLogicHandleMsg       = "RPCLogic.HandleMsg"
	RPCLogicVerifyUserToken = "RPCLogic.VerifyUserToken"
)

// VerifyUserToken verify app user token.
func (l *RPCLogic) VerifyUserToken(args *RPCLogicVerifyUserTokenArgs, reply *RPCLogicVerifyUserTokenReply) (err error) {
	return err
}

// HandleMsg handle user send msg.
func (l *RPCLogic) HandleMsg(args *RPCLogicHandleMsgArgs, reply *RPCLogicHandleMsgReply) (err error) {
	log.Println(RPCLogicHandleMsg, "is called:", args.User, args.Type, args.Msg)
	switch args.Type {
	case proto.PutMsg:
		reply.Msg, err = handleMsgMsg(args.User, args.Channel, args.Kind, args.Msg)
	default:
		return errors.New(ErrUnknownMsgType)
	}
	return err
}

func handleMsgMsg(user userboard.UserLocation, channel, kind string, msg interface{}) (replyMsg interface{}, err error) {
	// TODO
	// check org.user permission for channel.
	// errors.New(ErrPermDenied)
	if len(channel) < 3 {
		err = errors.New(ErrPermDenied)
		return
	}

	switch kind {
	case "":
		msgID, err := dispatcher.PutMsg(user, channel, msg)
		if err != nil {
			return nil, err
		}
		replyMsg = map[string]interface{}{
			"id": msgID,
		}
		return replyMsg, err
	case proto.PutStatusMsg:
		// channel status msg, eg. user typing.
		err := dispatcher.PutStatusMsg(user, channel, msg)
		return nil, err
	default:
		return nil, errors.New(ErrUnknownPutMsgKind)
	}
}
