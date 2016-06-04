package main

import (
	"log"

	"github.com/go-mangos/mangos"
	"github.com/go-mangos/mangos/protocol/rep"
	"github.com/go-mangos/mangos/protocol/req"
	"github.com/go-mangos/mangos/transport/ipc"
	"github.com/go-mangos/mangos/transport/tcp"
)

func startReqRepProxy() {
	sRep, err := rep.NewSocket()
	if err != nil {
		log.Fatal("failed to open reply socket:", err)
	}
	sRep.AddTransport(tcp.NewTransport())
	sRep.AddTransport(ipc.NewTransport())
	if err := sRep.Listen(args.repAddr); err != nil {
		log.Fatal("can't listen on reply socket:", err)
	}
	l.Info("reply listen on: %s", args.repAddr)

	sReq, err := req.NewSocket()
	if err != nil {
		log.Fatal("failed to open request socket:", err)
	}

	sReq.AddTransport(tcp.NewTransport())
	sReq.AddTransport(ipc.NewTransport())
	if err := sReq.Listen(args.reqAddr); err != nil {
		log.Fatal("can't listen on request socket:", err)
	}
	l.Info("request listen on: %s", args.reqAddr)

	if err := mangos.Device(sRep, sReq); err != nil {
		log.Fatal("start req/rep proxy error:", err)
	}
	l.Info("req/rep proxy started.")
}
