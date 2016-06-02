package main

import (
	"flag"
	"io"
	"runtime"
	"xim/utils/pprofutils"
	"xim/xchat/logic/service"

	"xim/xchat/logic/db"
	"xim/xchat/logic/logger"

	"github.com/valyala/gorpc"
)

var (
	l = logger.Logger
)

func onConnect(remoteAddr string, rwc io.ReadWriteCloser) (io.ReadWriteCloser, error) {
	l.Debug("client connected: [%s]", remoteAddr)
	return rwc, nil
}

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
	service.Init()

	d := service.NewServiceDispatcher()
	s := gorpc.NewTCPServer(args.addr, d.NewHandlerFunc())
	s.OnConnect = onConnect
	if err := s.Start(); err != nil {
		l.Critical("failed to start rpc service: [%s]", err)
	}
	defer s.Stop()

	setupSignal()
}
