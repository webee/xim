package broker

import (
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
func HandleLogicMsg(broker, org, user, instance string, msg []byte) ([]byte, error) {
	args := &rpcservice.RPCLogicHandleMsgArgs{
		Broker:   broker,
		Org:      org,
		User:     user,
		Instance: instance,
		Msg:      msg,
	}
	reply := new(rpcservice.RPCLogicHandleMsgReply)
	err := logicRPCClient.Client.Call(rpcservice.RPCLogicHandleMsg, args, reply)
	return reply.Msg, err
}
