package broker

import (
	"xim/logic"
	"xim/logic/rpcservice"
	"xim/utils/netutils"
	"xim/utils/rpcutils"
)

var (
	logicRPCClient *rpcutils.RPCClient
)

// InitLogicRPC connect to logic rpc.
func InitLogicRPC(netAddr *netutils.NetAddr) {
	logicRPCClient, _ = rpcutils.NewRPCClient(netAddr, true)
}

// HandleLogicMsg handle logic msg.
func HandleLogicMsg(user *logic.UserLocation, msgType string, channel string, msg interface{}) (interface{}, error) {
	args := &rpcservice.RPCLogicHandleMsgArgs{
		User:    *user,
		Type:    msgType,
		Channel: channel,
		Msg:     msg,
	}
	reply := new(rpcservice.RPCLogicHandleMsgReply)
	err := logicRPCClient.Client.Call(rpcservice.RPCLogicHandleMsg, args, reply)
	return reply.Msg, err
}
