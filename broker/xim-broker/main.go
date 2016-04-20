package main

import (
	"flag"
	"runtime"

	"xim/broker/userboard"
	"xim/utils/pprofutils"
)

var (
	userBoard = userboard.NewUserBaord()
)

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())

	if args.debug {
		pprofutils.StartPProfListen(args.pprofAddr)
	}
	//setupServer()
	initLogicRPC()
	startRPCService()
	startWebsocket()
	setupSignal()
}
