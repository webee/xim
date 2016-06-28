package main

import (
	"flag"
	"log"
	"runtime"
	"xim/xchat/broker/logger"

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

	// router
	router.Init()
	userKeys := map[string][]byte{
		"":     userKey,
		"test": testUserKey,
		"cs":   csUserKey,
	}
	xchatRouter, err := router.NewXChatRouter(userKeys, args.brokerDebug, args.testing, args.writeTimeout, args.idleTimeout)
	if err != nil {
		log.Fatalln("create xchat router failed:", err)
	}

	setupMid(xchatRouter)
	startRouter(xchatRouter)

	setupSignal()
}
