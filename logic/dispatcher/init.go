package dispatcher

import (
	"log"
	"xim/broker/userboard"
	"xim/dispatcher/rpcservice"
	"xim/utils/netutils"
	"xim/utils/rpcutils"
)

var (
	// TODO, 根据channel使用致性hash选择固定的dispatcher
	dispatcherRPCClient *rpcutils.RPCClient
)

// InitDispatcherRPC connect to dispatcher rpc.
func InitDispatcherRPC(netAddr *netutils.NetAddr) {
	dispatcherRPCClient, _ = rpcutils.NewRPCClient(netAddr, true)
}

// PutMsg push a msg to channel.
func PutMsg(user userboard.UserLocation, channel string, msg interface{}) (msgID string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	args := &rpcservice.RPCDispatcherPutMsgArgs{
		User:    user,
		Channel: channel,
		Msg:     msg,
	}
	log.Println("put:", user, channel, msg)
	reply := new(rpcservice.RPCDispatcherPutMsgReply)
	err = dispatcherRPCClient.Client.Call(rpcservice.RPCDispatcherPutMsg, args, reply)
	return reply.MsgID, err
}
