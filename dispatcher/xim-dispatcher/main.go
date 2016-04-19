package main

import (
	"flag"
	"xim/utils/pprofutils"
)

func main() {
	flag.Parse()

	if args.debug {
		pprofutils.StartPProfListen(args.pprofAddr)
	}
	startRPCService()
	setupSignal()
}
