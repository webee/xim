package main

import (
	"log"

	"github.com/go-mangos/mangos"
	"github.com/go-mangos/mangos/protocol/pub"
	"github.com/go-mangos/mangos/protocol/sub"
	"github.com/go-mangos/mangos/transport/ipc"
	"github.com/go-mangos/mangos/transport/tcp"
)

func startPubSubProxy() (close func()) {
	sPub, err := pub.NewSocket()
	if err != nil {
		log.Fatal("failed to open publish socket:", err)
	}
	sPub.AddTransport(tcp.NewTransport())
	sPub.AddTransport(ipc.NewTransport())
	for _, addr := range args.pubAddrs.List() {
		if err = sPub.Listen(addr); err != nil {
			log.Fatal("can't listen on publish socket:", err)
		}
		l.Info("publish listen on: %s", addr)
	}

	sSub, err := sub.NewSocket()
	if err != nil {
		log.Fatal("failed to open subscribe socket:", err)
	}
	sSub.AddTransport(tcp.NewTransport())
	sSub.AddTransport(ipc.NewTransport())
	for _, addr := range args.subAddrs.List() {
		if err = sSub.Listen(addr); err != nil {
			log.Fatal("can't listen on subscribe socket:", err)
		}
		l.Info("subscribe listen on: %s", addr)
	}
	err = sSub.SetOption(mangos.OptionSubscribe, []byte(""))
	if err != nil {
		log.Fatal("subscribe all messages error:", err)
	}

	if err := mangos.Device(sPub, sSub); err != nil {
		log.Fatal("start pub/sub proxy error:", err)
	}
	l.Info("pub/sub proxy started.")

	return func() {
		sPub.Close()
		sSub.Close()
	}
}
