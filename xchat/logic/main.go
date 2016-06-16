package main

import (
	"flag"
	"runtime"
	"xim/xchat/logic/logger"
	"xim/xchat/logic/pub"
	"xim/xchat/logic/service"

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

	defer db.InitDB(args.dbDriverName, args.dbDatasourceName, args.dbMaxConn)()
	defer pub.StartPublisher(args.pubAddrs.List(), args.dial)()
	defer nanorpc.StartRPCServer(args.addrs.List(), args.dial, new(service.RPCXChat))()

	setupSignal()
}
