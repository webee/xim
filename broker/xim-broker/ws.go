package main

import (
	"log"
	"xim/broker/ws"
)

func startWebsocket() {
	wsServer := ws.NewWebsocketServer(userBoard, &ws.UserServer{},
		ws.NewWebsocketServerConfig(
			&ws.WebsocketServerConfig{
				Testing:          args.testing,
				Addr:             args.addr,
				Broker:           args.rpcNetAddr,
				HTTPReadTimeout:  args.httpReadTimeout,
				HTTPWriteTimeout: args.httpWriteTimeout,
				HeartbeatTimeout: args.connHeartbeatTimeout,
			}))
	go func() {
		log.Fatal(wsServer.ListenAndServe())
	}()
}
