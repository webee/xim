package types

import "xim/broker/userds"

// RPCLogicHandleMsgArgs is the msg args.
type RPCLogicHandleMsgArgs struct {
	User    userds.UserLocation
	Type    string
	Channel string
	Kind    string
	Msg     interface{}
}

// RPCLogicHandleMsgReply is the msg reply.
type RPCLogicHandleMsgReply struct {
	Data interface{}
}

// RPCServer methods.
const (
	RPCLogicHandleMsg = "RPCLogic.HandleMsg"
)
