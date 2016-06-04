package main

import (
	"flag"
	"runtime"
	"xim/xchat/logic/logger"
	"xim/xchat/logic/rpcservice"

	"xim/utils/nanorpc"
	"xim/utils/pprofutils"

	"xim/xchat/logic/db"
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

	db.InitDB(args.dbDriverName, args.dbDatasourceName)
	nanorpc.StartRPCServer(args.addr, args.dial, new(rpcservice.RPCXChat))

	setupSignal()
}
