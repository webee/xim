package broker

import (
	"log"
	"xim/broker/proto"
	"xim/broker/rpcservice/types"
	"xim/broker/userds"
	"xim/utils/rpcutils"
)

// PushMsg push a msg to broker.
func PushMsg(user userds.UserLocation, msg *proto.Push) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	args := &types.RPCBrokerPushMsgArgs{
		User: user,
		Msg:  msg,
	}
	log.Println("push:", user, msg)
	reply := new(rpcutils.NoReply)

	// TODO handle error
	clientPool := rpcClientPoolCache.getRPCClientPool(user.Broker)
	id, client, _ := clientPool.Get()
	defer clientPool.Put(id)

	_ = client.Client.Go(types.RPCBrokerPushMsg, args, reply, nil)
	return nil
}
