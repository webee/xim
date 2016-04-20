package broker

import (
	"log"
	"xim/broker/proto"
	"xim/broker/rpcservice"
	"xim/logic"
	"xim/utils/netutils"
	"xim/utils/rpcutils"
)

// getBrokerRPCClient connect to dispatcher rpc.
func getBrokerRPCClient(broker string) *rpcutils.RPCClient {
	netAddr, _ := netutils.ParseNetAddr(broker)
	client, err := rpcutils.NewRPCClient(netAddr, true)
	if err == nil {
		return client
	}
	retryTimes := 3
	for retryTimes > 0 {
		client.Retry()
		if client.Connected() {
			break
		}
		retryTimes--
	}
	return client
}

// PushMsg push a msg to broker.
func PushMsg(user logic.UserLocation, msg proto.MsgMsg) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	args := &rpcservice.RPCBrokerPushMsgArgs{
		User: user,
		Msg:  msg,
	}
	log.Println("push:", user, msg)
	reply := new(rpcutils.NoReply)
	client := getBrokerRPCClient(user.Broker)
	err = client.Client.Call(rpcservice.RPCBrokerPushMsg, args, reply)
	log.Println("push result:", err)
	return err
}
