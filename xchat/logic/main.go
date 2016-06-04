package main

import (
	"flag"
	"net/rpc"
	"runtime"
	"xim/xchat/logic/logger"
	"xim/xchat/logic/rpcservice"

	"xim/utils/pprofutils"

	"xim/xchat/logic/db"
	"xim/xchat/logic/nanorpc"
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

	rpc.Register(new(rpcservice.RPCXChat))
	s := getReplySocket()
	go rpc.ServeCodec(nanorpc.NewNanoGobServerCodec(s))
	//startRPCServer()

	setupSignal()
}
