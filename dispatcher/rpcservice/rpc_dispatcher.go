package rpcservice

import (
	"log"
	"xim/broker/proto"
	"xim/broker/userdb"
	"xim/broker/userds"
	"xim/utils/rpcutils"
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
	User    userds.UserLocation
	Channel string
	Kind    string
	Msg     interface{}
}

// RPCDispatcherPutMsgReply is the msg reply.
type RPCDispatcherPutMsgReply struct {
	MsgID string
}

// RPCServer methods.
const (
	RPCDispatcherPutMsg       = "RPCDispatcher.PutMsg"
	RPCDispatcherPutStatusMsg = "RPCDispatcher.PutStatusMsg"
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

// PutStatusMsg put status msg to channel.
func (r *RPCDispatcher) PutStatusMsg(args *RPCDispatcherPutMsgArgs, reply *rpcutils.NoReply) error {
	log.Println(RPCDispatcherPutStatusMsg, "is called:", args.User, args.Channel, args.Msg)

	doDispatchStatusMsg(args.Channel, &args.User, args.Msg)
	return nil
}

func doDispatchMsg(channel string, user *userds.UserLocation, id, lastID string, msg interface{}) {
	log.Printf("dispatch msg: #%s, %s, [%s<-%s, %s]\n", channel, user, lastID, id, msg)
	protoMsg := &proto.ChannelMsg{
		Type:    proto.MsgMsg,
		Channel: channel,
		User:    user.User,
		ID:      id,
		LastID:  lastID,
		Msg:     msg,
	}
	putMsg(channel, user, protoMsg)
}

func doDispatchStatusMsg(channel string, user *userds.UserLocation, msg interface{}) {
	log.Printf("dispatch status msg: #%s, %s, %s\n", channel, user, msg)
	protoMsg := &proto.ChannelMsg{
		Type:    proto.MsgMsg,
		Channel: channel,
		User:    user.User,
		Kind:    proto.PutStatusMsg,
		Msg:     msg,
	}

	putMsg(channel, user, protoMsg)
}

func putMsg(channel string, user *userds.UserLocation, protoMsg *proto.ChannelMsg) {
	for _, u := range getChannelOnlineUserInstances(channel) {
		if *u == *user {
			continue
		}

		toPutMsg := &toPushMsg{
			user: *u,
			msg:  *protoMsg,
		}
		userMsgChan := userChannelCache.getMsgChan(u.String())
		userMsgChan.Put(toPutMsg)
	}
}

func getChannelOnlineUserInstances(channel string) []*userds.UserLocation {
	/*
		1. get channel users.
		2. filter get the online user instances.
	*/
	uids := []*userds.UserIdentity{
		&userds.UserIdentity{
			App:  "TEST",
			User: "webee",
		},
		&userds.UserIdentity{
			App:  "TEST",
			User: "xiaoee",
		},
	}
	if len(uids) == 0 {
		return []*userds.UserLocation{}
	}

	users, err := userdb.GetOnlineUsers(uids...)
	if err != nil {
		log.Printf("get online users error: channel->%s, %s\n", channel, err)
		return []*userds.UserLocation{}
	}

	return users
}
