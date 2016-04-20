package rpcservice

import (
	"fmt"
	"log"
	"math/rand"
	"xim/logic"

	"gopkg.in/redsync.v1"
)

// RPCDispatcher represents the rpc dispatcher.
type RPCDispatcher struct {
	redsync *redsync.Redsync
}

// NewRPCDispatcher returns a new rpc.
func NewRPCDispatcher(redsync *redsync.Redsync) *RPCDispatcher {
	return &RPCDispatcher{
		redsync: redsync,
	}
}

// RPCDispatcherPutMsgArgs is the msg args.
type RPCDispatcherPutMsgArgs struct {
	User    logic.UserLocation
	Channel string
	Msg     []byte
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
	log.Println(RPCDispatcherPutMsg, "is called:", args.User, args.Channel, string(args.Msg))
	reply.MsgID = fmt.Sprint(rand.Intn(100000))
	return err
}
