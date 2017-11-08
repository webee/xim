package main

import (
	"log"

	"github.com/go-mangos/mangos"
	"github.com/go-mangos/mangos/protocol/rep"
	"github.com/go-mangos/mangos/protocol/req"
	"github.com/go-mangos/mangos/transport/ipc"
	"github.com/go-mangos/mangos/transport/tcp"
	"time"
)

func startReqRepProxy() (close func()) {
	sRep, err := rep.NewSocket()
	if err != nil {
		log.Fatal("failed to open reply socket:", err)
	}
	sRep.AddTransport(tcp.NewTransport())
	sRep.AddTransport(ipc.NewTransport())
	for _, addr := range args.repAddrs.List() {
		if err = sRep.Listen(addr); err != nil {
			log.Fatal("can't listen on reply socket:", err)
		}
		l.Info("reply listen on: %s", addr)
	}

	sReq, err := req.NewSocket()
	if err != nil {
		log.Fatal("failed to open request socket:", err)
	}
	sReq.SetOption(mangos.OptionSendDeadline, 5*time.Second)
	// 不要重试
	sReq.SetOption(mangos.OptionRetryTime, 0)

	sReq.AddTransport(tcp.NewTransport())
	sReq.AddTransport(ipc.NewTransport())
	for _, addr := range args.reqAddrs.List() {
		if err = sReq.Listen(addr); err != nil {
			log.Fatal("can't listen on request socket:", err)
		}
		l.Info("request listen on: %s", addr)
	}

	if err := mangos.Device(sRep, sReq); err != nil {
		log.Fatal("start req/rep proxy error:", err)
	}
	l.Info("req/rep proxy started.")

	return func() {
		sRep.Close()
		sReq.Close()
	}
}
