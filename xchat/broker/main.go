package main

import (
	"xim/xchat/broker/logger"
	"xim/xchat/logic/service"

	"github.com/valyala/gorpc"

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

	// router
	router.Init()
	xchatRouter, err := router.NewXChatRouter(userKey, args.debug, args.testing)
	if err != nil {
		log.Fatalln("create xchat router failed:", err)
	}

	// xchat rpc service
	c := gorpc.NewTCPClient(args.logicRPCAddr)
	c.Start()
	defer c.Stop()
	d := service.NewServiceDispatcher()
	dc := d.NewServiceClient(service.XChat.Name, c)

	setupMid(xchatRouter, dc)
	startRouter(xchatRouter)
	setupSignal()
}
