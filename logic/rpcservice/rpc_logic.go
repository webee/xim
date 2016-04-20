package rpcservice

import (
	"encoding/json"
	"errors"
	"log"

	"xim/logic"
	"xim/logic/dispatcher"

	"github.com/bitly/go-simplejson"
)

// RPCLogic represents the rpc logic.
type RPCLogic struct {
}

// RPCLogicHandleMsgArgs is the msg args.
type RPCLogicHandleMsgArgs struct {
	User logic.UserLocation
	Type string
	Msg  json.RawMessage
}

// RPCLogicHandleMsgReply is the msg reply.
type RPCLogicHandleMsgReply struct {
	Msg json.RawMessage
}

// RPCServer methods.
const (
	MsgMsgType        = "msg"
	RPCLogicHandleMsg = "RPCLogic.HandleMsg"
)

// HandleMsg handle user send msg.
func (l *RPCLogic) HandleMsg(args *RPCLogicHandleMsgArgs, reply *RPCLogicHandleMsgReply) error {
	var (
		err error
	)
	log.Println(RPCLogicHandleMsg, "is called:", args.User, args.Type, string(args.Msg))
	switch args.Type {
	case MsgMsgType:
		reply.Msg, err = handleTextMsg(args.User, args.Msg)
		if err != nil {
			return errors.New(ErrHandlingMsg)
		}
	default:
		return errors.New(ErrUnknownMsgType)
	}
	return nil
}

func handleTextMsg(user logic.UserLocation, msg json.RawMessage) (replyMsg json.RawMessage, err error) {
	var (
		jd      *simplejson.Json
		msgID   string
		channel string
		txt     string
	)
	jd, err = simplejson.NewJson(msg)
	if err != nil {
		return nil, errors.New(ErrParseMsg)
	}
	channel, err = jd.Get("channel").String()
	if err != nil {
		return nil, errors.New(ErrBadMsg)
	}
	txt, err = jd.Get("txt").String()
	if err != nil || len(txt) == 0 {
		return nil, errors.New(ErrBadMsg)
	}
	// check org.user permission for channel.

	toSendMsg, _ := json.Marshal(&map[string]interface{}{
		"txt": txt,
	})
	msgID, err = dispatcher.PutMsg(user, channel, MsgMsgType, toSendMsg)
	if err != nil {
		return
	}
	replyMsg, err = json.Marshal(&map[string]interface{}{
		"txt": txt,
		"id":  msgID,
	})
	if err != nil {
		return nil, errors.New(ErrEncodingMsg)
	}
	return
}
