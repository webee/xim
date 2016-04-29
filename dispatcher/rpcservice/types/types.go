package types

import "xim/broker/userds"

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
