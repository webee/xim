package rpcservice

import (
	"log"
	"xim/broker/rpcservice/types"
	"xim/broker/userboard"
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

// PushMsg push msg to broker.
func (r *RPCBroker) PushMsg(args *types.RPCBrokerPushMsgArgs, reply *rpcutils.NoReply) error {
	log.Println("GET PUSH:", args.User, args.Msg)
	userConn, err := r.userBoard.GetUserConn(&args.User)
	if err != nil {
		return err
	}
	return userConn.PushMsg(&args.Msg)
}
