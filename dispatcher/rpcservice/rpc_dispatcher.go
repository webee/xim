package rpcservice

import (
	"encoding/json"
	"log"
	"xim/broker/proto"
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

// PutMsg put msg to channel.
func (r *RPCDispatcher) PutMsg(args *RPCDispatcherPutMsgArgs, reply *RPCDispatcherPutMsgReply) error {
	var err error
	log.Println(RPCDispatcherPutMsg, "is called:", args.User, args.Channel, args.Type, string(args.Msg))
	msgChan := channelCache.getMsgChan(args.Channel)
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
	protoMsg := proto.MsgMsg{
		Type:    msgType,
		Channel: channel,
		User:    user.User,
		ID:      id,
		LastID:  lastID,
		Msg:     msg.(json.RawMessage),
	}
	for _, user := range getChannelOnlineUserInstances(channel) {
		toPutMsg := &toPushMsg{
			user: user,
			msg:  protoMsg,
		}
		userMsgChan := userChannelCache.getMsgChan(user.String())
		userMsgChan.Put(toPutMsg)
	}
}

func getChannelOnlineUserInstances(channel string) []logic.UserLocation {
	/*
		1. get channel users.
		2. filter get the online user instances.
	*/
	return []logic.UserLocation{
		logic.UserLocation{
			Broker:   "tcp@localhost:5780",
			Org:      "test",
			User:     "webee",
			Instance: "1",
		},
		logic.UserLocation{
			Broker:   "tcp@localhost:5780",
			Org:      "test",
			User:     "xiaoee",
			Instance: "2",
		},
	}
}
