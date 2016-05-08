package main

import (
	"log"
	"xim/broker/ws"
)

func startAppWebsocket() {
	wsServer := ws.NewWebsocketServer(userBoard, ws.NewAppServer(),
		ws.NewWebsocketServerConfig(
			&ws.WebsocketServerConfig{
				Testing:          args.testing,
				Addr:             args.appAddr,
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
