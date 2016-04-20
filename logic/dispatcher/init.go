package dispatcher

import (
	"encoding/json"
	"xim/dispatcher/rpcservice"
	"xim/logic"
	"xim/utils/netutils"
	"xim/utils/rpcutils"
)

var (
	dispatcherRPCClient *rpcutils.RPCClient
)

// InitDispatcherRPC connect to dispatcher rpc.
func InitDispatcherRPC(netAddr *netutils.NetAddr) {
	dispatcherRPCClient, _ = rpcutils.NewRPCClient(netAddr, true)
}

// PutMsg push a msg to channel.
func PutMsg(from logic.UserLocation, channel string, msg json.RawMessage) (msgID string, err error) {
	args := &rpcservice.RPCDispatcherPutMsgArgs{
		User:    from,
		Channel: channel,
		Msg:     msg,
	}
	reply := new(rpcservice.RPCDispatcherPutMsgReply)
	err = dispatcherRPCClient.Client.Call(rpcservice.RPCDispatcherPutMsg, args, reply)
	return reply.MsgID, err
}
