package main

import (
	"flag"
	"runtime"
	"xim/utils/pprofutils"
)

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())

	if args.debug {
		pprofutils.StartPProfListen(args.pprofAddr)
	}
	initDB()
	//startHTTPApiServer()
	initDispatcherRPC()
	startRPCService()
	setupSignal()
}
