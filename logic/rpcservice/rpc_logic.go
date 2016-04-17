package rpcservice

import (
	"encoding/json"
	"log"
)

// RPCLogic represents the rpc logic.
type RPCLogic struct {
}

// RPCLogicHandleMsgArgs is the msg args.
type RPCLogicHandleMsgArgs struct {
	Broker   string
	Org      string
	User     string
	Instance string
	Msg      []byte
}

// RPCLogicHandleMsgReply is the msg reply.
type RPCLogicHandleMsgReply struct {
	Msg []byte
}

// RPCServer methods.
const (
	RPCLogicHandleMsg = "RPCLogic.HandleMsg"
)

// HandleMsg handle user send msg.
func (l *RPCLogic) HandleMsg(args *RPCLogicHandleMsgArgs, reply *RPCLogicHandleMsgReply) error {
	var err error
	log.Println(RPCLogicHandleMsg, "is called:", args.Broker, args.Org, args.User, args.Instance, string(args.Msg))
	reply.Msg, err = json.Marshal(&map[string]interface{}{
		"ok": true,
	})
	return err
}
