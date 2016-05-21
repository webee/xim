package logic

import (
	"xim/broker/userds"
	"xim/logic/rpcservice/types"
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

// PutMsg put msg.
func PutMsg(user *userds.UserLocation, channel string, msgKind string, msg interface{}) (interface{}, error) {
	args := &types.RPCLogicPutMsgArgs{
		User:    *user,
		Channel: channel,
		Kind:    msgKind,
		Msg:     msg,
	}
	reply := new(types.RPCLogicPutMsgReply)
	err := logicRPCClient.Client.Call(types.RPCLogicPutMsg, args, reply)
	return reply.Data, err
}
