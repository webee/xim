package rpcservice

import (
	"errors"
	"log"
	"xim/broker/proto"
	"xim/broker/userdb"
	"xim/broker/userds"
	"xim/commons/db"
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
	msgChan := channelCache.getMsgChan(args.Channel)

	qm := &queueMsg{
		user:    args.User,
		channel: args.Channel,
		msg:     args.Msg,
		id:      make(chan int, 1),
		ts:      make(chan int64, 1),
	}
	log.Println("sending", args.User, args.Channel, args.Msg)
	if err = msgChan.Put(qm); err == nil {
		var open bool
		reply.MsgID, open = <-qm.id
		if !open {
			return errors.New("send failed")
		}
		reply.Timestamp, open = <-qm.ts
		if !open {
			return errors.New("send failed")
		}
	}
	return err
}

// PutStatusMsg put status msg to channel.
func (r *RPCDispatcher) PutStatusMsg(args *types.RPCDispatcherPutMsgArgs, reply *rpcutils.NoReply) error {
	log.Println(types.RPCDispatcherPutStatusMsg, "is called:", args.User, args.Channel, args.Msg)

	// FIXME: 是否考虑状态消息顺序?
	doDispatchMsg(args.Channel, &args.User, nil, proto.PutStatusMsgKind, args.Msg, nil)
	return nil
}

func doDispatchMsg(channel string, user *userds.UserLocation, id interface{}, kind string, msg interface{}, ts interface{}) {
	log.Printf("dispatch %s msg: #%s, %s, [%d, %s]\n", kind, channel, user, id, msg)
	protoMsg := &proto.Push{
		Channel: channel,
		User:    user.User,
		ID:      uint64(id.(int)),
		Kind:    kind,
		Msg:     msg,
		Ts:      uint64(ts.(int64)),
	}
	putMsg(channel, user, protoMsg)
}

func putMsg(channel string, user *userds.UserLocation, protoMsg *proto.Push) {
	for _, u := range getChannelOnlineUserInstances(user.App, channel) {
		if *u == *user {
			continue
		}

		toPutMsg := &toPushMsg{
			user: *u,
			msg:  protoMsg,
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
