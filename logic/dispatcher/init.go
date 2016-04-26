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
	// TODO, 连接池
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

// PutStatusMsg push a msg to channel.
func PutStatusMsg(user userboard.UserLocation, channel string, msg interface{}) (err error) {
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
	reply := new(rpcutils.NoReply)
	err = dispatcherRPCClient.Client.Call(rpcservice.RPCDispatcherPutStatusMsg, args, reply)
	return err
}
