package rpcservice

import (
	"encoding/json"
	"log"
	"xim/broker/proto"
	"xim/dispatcher/broker"
	"xim/logic"
)

// RPCDispatcher represents the rpc dispatcher.
type RPCDispatcher struct {
}

// NewRPCDispatcher returns a new rpc.
func NewRPCDispatcher() *RPCDispatcher {
	return &RPCDispatcher{}
}

// RPCDispatcherPutMsgArgs is the msg args.
type RPCDispatcherPutMsgArgs struct {
	User    logic.UserLocation
	Channel string
	Type    string
	Msg     json.RawMessage
}

// RPCDispatcherPutMsgReply is the msg reply.
type RPCDispatcherPutMsgReply struct {
	MsgID string
}

// RPCServer methods.
const (
	RPCDispatcherPutMsg = "RPCDispatcher.PutMsg"
)

var ()

// PutMsg put msg to channel.
func (r *RPCDispatcher) PutMsg(args *RPCDispatcherPutMsgArgs, reply *RPCDispatcherPutMsgReply) error {
	var err error
	log.Println(RPCDispatcherPutMsg, "is called:", args.User, args.Channel, args.Type, string(args.Msg))
	msgChan := channels.getMsgChan(args.Channel)
	qm := &queueMsg{
		user:    args.User,
		channel: args.Channel,
		msgType: args.Type,
		msg:     args.Msg,
		id:      make(chan string, 1),
	}
	if err = msgChan.Put(qm); err == nil {
		reply.MsgID = <-qm.id
	}
	return err
}

func doDispatchMsg(channel string, user logic.UserLocation, msgType, id, lastID string, msg interface{}) {
	log.Printf("dispatch msg: #%s, %s, [%s<-%s, %s]\n", channel, user, lastID, id, msg)
	/*
		type: msg
		channel: xxx
		user: "webee"
		id: 1461145447.1
		last_id: 1461145446.2
		msg: "xxxx"
	*/
	/*
		1. getChannelOnlineUserInstances.
		2. for each user instance, dispatch the msg.

		msg:
			org: "AAA"
			user: "webee"
			instance: "#1"
			msg: {
			// ref upper.
			}
	*/
	toSendMsg := proto.MsgMsg{
		Type:    msgType,
		Channel: channel,
		User:    user.User,
		ID:      id,
		LastID:  lastID,
		Msg:     msg.(json.RawMessage),
	}
	broker.PushMsg(user, toSendMsg)
}
