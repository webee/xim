package main

import (
	"log"
	"xim/broker/ws"
)

func startWebsocket() {
	wsServer := ws.NewWebsocketServer(userBoard, ws.NewWebsocketServerConfig(
		&ws.WebsocketServerConfig{
			Testing:          args.testing,
			Addr:             args.addr,
			Broker:           args.rpcNetAddr,
			HTTPReadTimeout:  args.httpReadTimeout,
			HTTPWriteTimeout: args.httpWriteTimeout,
			HeartbeatTimeout: args.connHeartbeatTimeout,
			WriteTimeout:     args.connWriteTimeout,
		}))
	go func() {
		log.Fatal(wsServer.ListenAndServe())
	}()
}
