package main

import (
	"log"
	"xim/logic/dispatcher"
	"xim/utils/netutils"
)

// initDispatcherRPC: connect to logic rpc.
func initDispatcherRPC() {
	netAddr, err := netutils.ParseNetAddr(args.dispatcherRPCNetAddr)
	if err != nil {
		log.Fatalln(args.dispatcherRPCNetAddr, err)
	}
	dispatcher.InitDispatcherRPC(netAddr)
}
