package main

import (
	"flag"
	"runtime"
	"xim/utils/pprofutils"
	"xim/xchat/proxy/logger"
)

var (
	l = logger.Logger
)

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())

	if args.debug {
		pprofutils.StartPProfListen(args.pprofAddr)
	}

	startPubSubProxy()
	startReqRepProxy()

	setupSignal()
}
