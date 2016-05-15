package main

import (
	"flag"
	"log"
	"runtime"
	"xim/apps/xchat/router"
	"xim/utils/pprofutils"
)

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())

	if args.debug {
		pprofutils.StartPProfListen(args.pprofAddr)
	}
	setupKeys()
	initDB()

	xchatRouter, err := router.NewXChatRouter(userKey, args.debug, args.testing)
	if err != nil {
		log.Fatalln("create xchat router failed:", err)
	}

	setupMid(xchatRouter)

	startRouter(xchatRouter)
	setupSignal()
}
