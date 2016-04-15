package main

import (
	"log"
	"xim/utils/netutils"
	"xim/utils/rpcutils"
)

var (
	logicRPCClient *rpcutils.RPCClient
)

// InitLogicRPC: connect to logic rpc.
func initLogicRPC() {
	netAddr, err := netutils.ParseNetAddr(args.logicRPCNetAddr)
	if err != nil {
		log.Fatalln(args.logicRPCNetAddr, err)
	}
	logicRPCClient, _ = rpcutils.NewRPCClient(netAddr, true)
}
