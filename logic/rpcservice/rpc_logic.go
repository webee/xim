package rpcservice

import (
	"errors"
	"log"

	"xim/broker/proto"
	"xim/broker/userds"
	"xim/logic/db"
	"xim/logic/dispatcher"
	"xim/logic/rpcservice/types"
)

// RPCLogic represents the rpc logic.
type RPCLogic struct {
}

// HandleMsg handle user send msg.
func (l *RPCLogic) HandleMsg(args *types.RPCLogicHandleMsgArgs, reply *types.RPCLogicHandleMsgReply) (err error) {
	log.Println(types.RPCLogicHandleMsg, "is called:", args.User, args.Type, args.Msg)
	switch args.Type {
	case proto.PutMsg:
		reply.Msg, err = handleMsgMsg(args.User, args.Channel, args.Kind, args.Msg)
	default:
		return errors.New(ErrUnknownMsgType)
	}
	return err
}

func handleMsgMsg(user userds.UserLocation, channel, kind string, msg interface{}) (replyMsg interface{}, err error) {
	if !db.CanUserPubChannel(user, channel) {
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
	case proto.PutNotifyMsg:
		// channel notify msg
		err := dispatcher.PutStatusMsg(user, channel, msg)
		return nil, err
	default:
		return nil, errors.New(ErrUnknownPutMsgKind)
	}
}
