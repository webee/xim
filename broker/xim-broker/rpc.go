package main

import (
	"log"
	"xim/broker/rpcservice"
	"xim/utils/netutils"
)

func startRPCService() {
	netAddr, err := netutils.ParseNetAddr(args.rpcNetAddr)
	if err != nil {
		log.Fatalln(args.rpcNetAddr, err)
	}

	rpcservice.StartRPCServer(netAddr,
		rpcservice.NewRPCBroker(userBoard),
	)
}
