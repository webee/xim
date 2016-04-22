package rpcservice

import (
	"log"
	"xim/broker/proto"
	"xim/broker/userboard"
	"xim/logic"
	"xim/utils/rpcutils"
)

// RPCBroker represents the rpc broker.
type RPCBroker struct {
	userBoard *userboard.UserBoard
}

// NewRPCBroker returns a new rpc.
func NewRPCBroker(userBoard *userboard.UserBoard) *RPCBroker {
	return &RPCBroker{userBoard}
}

// RPCBrokerPushMsgArgs is the msg args.
type RPCBrokerPushMsgArgs struct {
	User logic.UserLocation
	Msg  proto.MsgMsg
}

// RPCServer methods.
const (
	RPCBrokerPushMsg = "RPCBroker.PushMsg"
)

// PushMsg push msg to broker.
func (r *RPCBroker) PushMsg(args *RPCBrokerPushMsgArgs, reply *rpcutils.NoReply) error {
	log.Println("GET PUSH:", args.User, args.Msg)
	uid := userboard.NewUserIdentify(args.User.Org, args.User.User)
	userConn, err := r.userBoard.GetUserConn(uid, args.User.Instance)
	if err != nil {
		return err
	}
	return userConn.PushMsg(&args.Msg)
}
