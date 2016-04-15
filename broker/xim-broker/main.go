package main

import (
	"flag"
	"runtime"

	"xim/broker"
	"xim/utils/pprofutils"
)

var (
	userBoard = broker.NewUserBaord()
)

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())

	if args.debug {
		pprofutils.StartPProfListen(args.pprofAddr)
	}
	//setupServer()
	initLogicRPC()
	startWebsocket()
	setupSignal()
}
