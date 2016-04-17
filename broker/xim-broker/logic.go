package main

import (
	"log"
	"xim/broker"
	"xim/utils/netutils"
)

// initLogicRPC: connect to logic rpc.
func initLogicRPC() {
	netAddr, err := netutils.ParseNetAddr(args.logicRPCNetAddr)
	if err != nil {
		log.Fatalln(args.logicRPCNetAddr, err)
	}
	broker.InitLogicRPC(netAddr)
}
