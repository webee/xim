package main

import (
	"log"
	"xim/broker"
)

func startWebsocket() {
	wsServer := broker.NewWebsocketServer(userBoard, broker.NewWebsocketServerConfig(
		&broker.WebsocketServerConfig{
			Testing: args.testing,
			Addr:    args.addr,
			Broker:  args.rpcNetAddr,
		}))
	go func() {
		log.Fatal(wsServer.ListenAndServe())
	}()
}
