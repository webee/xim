package main

import (
	"flag"
	"log"
	"runtime"
	"xim/utils/pprofutils"
	"xim/xchat/broker/router"
)

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())

	if args.debug {
		pprofutils.StartPProfListen(args.pprofAddr)
	}

	setupKeys()

	xchatRouter, err := router.NewXChatRouter(userKey, args.debug, args.testing)
	if err != nil {
		log.Fatalln("create xchat router failed:", err)
	}

	setupMid(xchatRouter)

	startRouter(xchatRouter)
	setupSignal()
}
