package main

import (
	"xim/xchat/broker/logger"

	"flag"
	"log"
	"runtime"
	"xim/utils/pprofutils"
	"xim/xchat/broker/router"
)

var (
	l = logger.Logger
)

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())

	if !args.debug {
		l.MaxLevel = 6
	}
	defer l.Close()

	if args.debug {
		pprofutils.StartPProfListen(args.pprofAddr)
	}

	setupKeys()

	router.Init()
	xchatRouter, err := router.NewXChatRouter(userKey, args.debug, args.testing)
	if err != nil {
		log.Fatalln("create xchat router failed:", err)
	}

	setupMid(xchatRouter)

	startRouter(xchatRouter)
	setupSignal()
}
