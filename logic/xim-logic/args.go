package main

import "flag"

// Args is app's arguments.
type Args struct {
	rpcNetAddr string
	testing    bool
	debug      bool
	pprofAddr  string
}

var (
	args Args
)

func init() {
	flag.StringVar(&args.rpcNetAddr, "rpc-net-addr", "tcp@localhost:6780", "rpc network address to listen.")
	flag.BoolVar(&args.testing, "testing", false, "whether to serv a testing page.")
	flag.BoolVar(&args.debug, "debug", false, "whether to enable debug tools.")
	flag.StringVar(&args.pprofAddr, "pprof-addr", "localhost:6060", "debug pprof http address.")
}
