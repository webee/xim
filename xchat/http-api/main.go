package main

import (
	"flag"
	"runtime"
	"xim/xchat/broker/logger"
	"xim/xchat/http-api/server"

	"xim/utils/pprofutils"
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
	userKeys := map[string][]byte{
		"":       userKey,
		"test":   testUserKey,
		"cs":     csUserKey,
		"notify": notifyUserKey,
	}

	server.Start(&server.Config{
		Debug:          args.debug,
		Testing:        args.testing,
		Keys:           userKeys,
		Addr:           args.addr,
		LogicRPCAddr:   args.logicRPCAddr,
		RPCCallTimeout: args.rpcCallTimeout,
		TurnUser:       args.turnUser,
		TurnSecret:     args.turnSecret,
		TurnURI:        args.turnURI,
	})
}
