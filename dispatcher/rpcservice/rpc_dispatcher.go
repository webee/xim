package rpcservice

import (
	"log"
	"xim/broker/proto"
	"xim/broker/userboard"
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
	User    userboard.UserLocation
	Channel string
	Msg     interface{}
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
	log.Println(RPCDispatcherPutMsg, "is called:", args.User, args.Channel, args.Msg)
	msgChan := channelCache.getMsgChan(args.Channel)
	qm := &queueMsg{
		user:    args.User,
		channel: args.Channel,
		msg:     args.Msg,
		id:      make(chan string, 1),
	}
	if err = msgChan.Put(qm); err == nil {
		reply.MsgID = <-qm.id
	}
	return err
}

func doDispatchMsg(channel string, user userboard.UserLocation, id, lastID string, msg interface{}) {
	log.Printf("dispatch msg: #%s, %s, [%s<-%s, %s]\n", channel, user, lastID, id, msg)
	protoMsg := proto.MsgMsg{
		Channel: channel,
		User:    user.User,
		ID:      id,
		LastID:  lastID,
		Msg:     msg,
	}
	for _, user := range getChannelOnlineUserInstances(channel) {
		toPutMsg := &toPushMsg{
			user: *user,
			msg:  protoMsg,
		}
		userMsgChan := userChannelCache.getMsgChan(user.String())
		userMsgChan.Put(toPutMsg)
	}
}

func getChannelOnlineUserInstances(channel string) []*userboard.UserLocation {
	/*
		1. get channel users.
		2. filter get the online user instances.
	*/
	uids := []*userboard.UserIdentity{
		&userboard.UserIdentity{
			Org:  "test",
			User: "webee",
		},
		&userboard.UserIdentity{
			Org:  "test",
			User: "xiaoee",
		},
	}
	if len(uids) == 0 {
		return []*userboard.UserLocation{}
	}

	users, err := userboard.GetOnlineUsers(uids...)
	if err != nil {
		log.Printf("get online users error: channel->%s, %s\n", channel, err)
		return []*userboard.UserLocation{}
	}

	return users
}
