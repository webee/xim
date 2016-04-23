package broker

import (
	"log"
	"xim/broker/proto"
	"xim/broker/rpcservice"
	"xim/logic"
	"xim/utils/rpcutils"
)

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

	// TODO handle error
	clientPool := rpcClientPoolCache.getRPCClientPool(user.Broker)
	id, client, _ := clientPool.Get()
	defer clientPool.Put(id)

	err = client.Client.Call(rpcservice.RPCBrokerPushMsg, args, reply)
	if err != nil {
		log.Println("push err:", err)
		client.Reconnect()
		err = client.Client.Call(rpcservice.RPCBrokerPushMsg, args, reply)
		if err != nil {
			log.Println("retry push err:", err)
		}
	}
	return err
}
