package types

import "xim/broker/userds"

// RPCLogicPutMsgArgs is the msg args.
type RPCLogicPutMsgArgs struct {
	User    userds.UserLocation
	Channel string
	Kind    string
	Msg     interface{}
}

// RPCLogicPutMsgReply is the msg reply.
type RPCLogicPutMsgReply struct {
	Data interface{}
}

// RPCServer methods.
const (
	RPCLogicPutMsg = "RPCLogic.PutMsg"
)
