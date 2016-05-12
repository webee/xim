package rpcservice

import (
	"log"
	"xim/broker/proto"
	"xim/broker/userdb"
	"xim/broker/userds"
	"xim/dispatcher/db"
	"xim/dispatcher/rpcservice/types"
	"xim/utils/rpcutils"
)

// RPCDispatcher represents the rpc dispatcher.
type RPCDispatcher struct {
}

// NewRPCDispatcher returns a new rpc.
func NewRPCDispatcher() *RPCDispatcher {
	return &RPCDispatcher{}
}

// PutMsg put msg to channel.
func (r *RPCDispatcher) PutMsg(args *types.RPCDispatcherPutMsgArgs, reply *types.RPCDispatcherPutMsgReply) error {
	var err error
	log.Println(types.RPCDispatcherPutMsg, "is called:", args.User, args.Channel, args.Msg)
	// FIXME: channelCache在大量请求时无限阻塞
	msgChan := channelCache.getMsgChan(args.Channel)
	qm := &queueMsg{
		user:    args.User,
		channel: args.Channel,
		msg:     args.Msg,
		id:      make(chan string, 1),
	}
	log.Println("sending", args.User, args.Channel, args.Msg)
	if err = msgChan.Put(qm); err == nil {
		reply.MsgID = <-qm.id
	}
	return err
}

// PutStatusMsg put status msg to channel.
func (r *RPCDispatcher) PutStatusMsg(args *types.RPCDispatcherPutMsgArgs, reply *rpcutils.NoReply) error {
	log.Println(types.RPCDispatcherPutStatusMsg, "is called:", args.User, args.Channel, args.Msg)

	doDispatchMsg(args.Channel, &args.User, "", "", proto.PutStatusMsg, args.Msg)
	return nil
}

func doDispatchMsg(channel string, user *userds.UserLocation, id, lastID string, kind string, msg interface{}) {
	log.Printf("dispatch %s msg: #%s, %s, [%s<-%s, %s]\n", kind, channel, user, lastID, id, msg)
	protoMsg := &proto.ChannelMsg{
		Type:    proto.MsgMsg,
		Channel: channel,
		User:    user.User,
		ID:      id,
		LastID:  lastID,
		Kind:    kind,
		Msg:     msg,
	}
	putMsg(channel, user, protoMsg)
}

func putMsg(channel string, user *userds.UserLocation, protoMsg *proto.ChannelMsg) {
	for _, u := range getChannelOnlineUserInstances(user.App, channel) {
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

func getChannelOnlineUserInstances(app, channel string) []*userds.UserLocation {
	uids := db.GetChannelSubscribers(app, channel)
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
