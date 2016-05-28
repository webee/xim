package main

import (
	"flag"
	"runtime"

	"xim/broker/userboard"
	"xim/utils/pprofutils"
)

var (
	userBoard = userboard.NewUserBoard()
)

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())

	if args.debug {
		pprofutils.StartPProfListen(args.pprofAddr)
	}
	initUserboard()
	initLogicRPC()
	startRPCService()
	initWebsocket()
	setupSignal()
}
