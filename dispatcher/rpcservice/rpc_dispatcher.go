package rpcservice

import (
	"log"
	"math/rand"
)

// RPCDispatcher represents the rpc dispatcher.
type RPCDispatcher struct {
}

// RPCDispatcherPutMsgArgs is the msg args.
type RPCDispatcherPutMsgArgs struct {
	Broker   string
	Org      string
	User     string
	Instance string
	Channel  string
	Msg      []byte
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
	log.Println(RPCDispatcherPutMsg, "is called:", args)
	reply.MsgID = string(rand.Intn(100000))
	return err
}
