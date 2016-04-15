package main

import (
	"flag"
	"runtime"

	"xim/broker"
)

var (
	userBoard = broker.NewUserBaord()
)

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())

	//setupServer()
	startWebsocket()
	setupSignal()
}
