package main

import (
	"log"
	"xim/dispatcher"
	"xim/dispatcher/rpcservice"
	"xim/utils/netutils"

	"gopkg.in/redsync.v1"
)

func startRPCService() {
	netAddr, err := netutils.ParseNetAddr(args.rpcNetAddr)
	if err != nil {
		log.Fatalln(args.rpcNetAddr, err)
	}

	redisPool := dispatcher.NewRedisPool(args.redisServer, args.redisPassword)
	rpcservice.StartRPCServer(netAddr,
		rpcservice.NewRPCDispatcher(redsync.New([]redsync.Pool{redisPool})),
	)
}
