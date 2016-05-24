package rpcservice

import (
	"errors"
	"log"

	"xim/broker/proto"
	"xim/broker/userds"
	"xim/commons/db"
	"xim/logic/dispatcher"
	"xim/logic/rpcservice/types"
)

// RPCLogic represents the rpc logic.
type RPCLogic struct {
}

// PutMsg put user send msg to channel.
func (l *RPCLogic) PutMsg(args *types.RPCLogicPutMsgArgs, reply *types.RPCLogicPutMsgReply) (err error) {
	log.Println(types.RPCLogicPutMsg, "is called:", args.User, args.Channel, args.Msg)
	reply.Data, err = handlePutMsg(args.User, args.Channel, args.Kind, args.Msg)
	return err
}

func handlePutMsg(user userds.UserLocation, channel, kind string, msg interface{}) (replyData interface{}, err error) {
	if !db.CanUserPubChannel(user, channel) {
		err = errors.New(ErrPermDenied)
		return
	}

	switch kind {
	case "":
		msgID, ts, err := dispatcher.PutMsg(user, channel, msg)
		if err != nil {
			return nil, err
		}
		replyData = map[string]interface{}{
			"id": msgID,
			"ts": ts,
		}
		return replyData, err
	case proto.PutStatusMsgKind:
		// channel status msg, eg. user typing.
		err := dispatcher.PutStatusMsg(user, channel, msg)
		return nil, err
	case proto.PutNotifyMsgKind:
		// channel notify msg
		err := dispatcher.PutStatusMsg(user, channel, msg)
		return nil, err
	default:
		return nil, errors.New(ErrUnknownPutMsgKind)
	}
}
