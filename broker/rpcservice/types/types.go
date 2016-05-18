package types

import (
	"xim/broker/proto"
	"xim/broker/userds"
)

// RPCBrokerPushMsgArgs is the msg args.
type RPCBrokerPushMsgArgs struct {
	User userds.UserLocation
	Msg  *proto.Push
}

// RPCServer methods.
const (
	RPCBrokerPushMsg = "RPCBroker.PushMsg"
)
