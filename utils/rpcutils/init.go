package rpcutils

import (
	"log"
	"net"
	"net/rpc"
)

// RegisterAndStartRPCServer register services and start listen.
func RegisterAndStartRPCServer(network, laddr string, rcvrs ...interface{}) {
	for _, rcvr := range rcvrs {
		rpc.Register(rcvr)
	}
	go rpcListen(network, laddr)
}

func rpcListen(network, laddr string) {
	l, err := net.Listen(network, laddr)
	if err != nil {
		log.Fatalf("net.Listen(%q, %q) error(%v)\n", network, laddr, err)
	}
	// if process exit, then close the rpc bind
	defer func() {
		if err := l.Close(); err != nil {
			log.Panicf("listener.Close() error(%v).\n", err)
		}
		log.Printf("listen %q %q close.\n", network, laddr)
	}()
	log.Printf("rpc listen %s@%s.\n", network, laddr)
	rpc.Accept(l)
}
