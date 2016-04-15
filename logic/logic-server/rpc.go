package main

import (
	"log"
	"xim/logic/rpcservice"
	"xim/utils/netutils"
)

func startRPCService() {
	netAddr, err := netutils.ParseNetAddr(args.rpcNetAddr)
	if err != nil {
		log.Fatalln(args.rpcNetAddr, err)
	}

	rpcservice.StartRPCServer(netAddr)
}
