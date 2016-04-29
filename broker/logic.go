package broker

import (
	"xim/broker/userds"
	"xim/logic/rpcservice"
	"xim/utils/netutils"
	"xim/utils/rpcutils"
)

var (
	logicRPCClient *rpcutils.RPCClient
)

// InitLogicRPC connect to logic rpc.
func InitLogicRPC(netAddr *netutils.NetAddr) {
	// TODO use connection pool.
	logicRPCClient, _ = rpcutils.NewRPCClient(netAddr, true)
}

// HandleLogicMsg handle logic msg.
func HandleLogicMsg(user *userds.UserLocation, msgType string, channel string, msgKind string, msg interface{}) (interface{}, error) {
	args := &rpcservice.RPCLogicHandleMsgArgs{
		User:    *user,
		Type:    msgType,
		Channel: channel,
		Kind:    msgKind,
		Msg:     msg,
	}
	reply := new(rpcservice.RPCLogicHandleMsgReply)
	err := logicRPCClient.Client.Call(rpcservice.RPCLogicHandleMsg, args, reply)
	return reply.Msg, err
}
