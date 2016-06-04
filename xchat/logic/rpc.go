package main

import (
	"log"

	"github.com/go-mangos/mangos"
	"github.com/go-mangos/mangos/protocol/rep"
	"github.com/go-mangos/mangos/transport/ipc"
	"github.com/go-mangos/mangos/transport/tcp"
)

func getReplySocket() mangos.Socket {
	s, err := rep.NewSocket()
	if err != nil {
		log.Fatal("failed to open reply socket:", err)
	}

	s.SetOption(mangos.OptionRaw, true)
	s.AddTransport(tcp.NewTransport())
	s.AddTransport(ipc.NewTransport())

	if args.dial {
		// dial to load balancing rep/req proxy.
		if err := s.Dial(args.addr); err != nil {
			log.Fatal("can't dial on request socket:", err)
		}
		l.Info("rpc dial to: %s", args.addr)
	} else {
		// serve
		if err := s.Listen(args.addr); err != nil {
			log.Fatal("can't listen on reply socket:", err)
		}
		l.Info("rpc listen on: %s", args.addr)
	}
	return s
}
