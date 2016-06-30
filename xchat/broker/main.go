package main

import (
	"flag"
	"log"
	"runtime"
	"xim/xchat/broker/logger"

	"xim/utils/pprofutils"
	"xim/xchat/broker/mid"
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
	xchatRouter, err := router.NewXChatRouter(userKeys, args.brokerDebug, args.testing, args.writeTimeout, args.pingTimeout, args.idleTimeout)
	if err != nil {
		log.Fatalln("create xchat router failed:", err)
	}

	mid.Setup(&mid.Config{
		Debug:          args.debug,
		Testing:        args.testing,
		Key:            userKey,
		LogicRPCAddr:   args.logicRPCAddr,
		LogicPubAddr:   args.logicPubAddr,
		XChatHostURL:   args.xchatHostURL,
		RPCCallTimeout: args.rpcCallTimeout,
	}, xchatRouter)

	startRouter(xchatRouter)

	setupSignal()
}
