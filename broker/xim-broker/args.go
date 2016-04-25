package main

import "flag"

// Args is app's arguments.
type Args struct {
	addr            string
	testing         bool
	debug           bool
	logicRPCNetAddr string
	pprofAddr       string
	rpcNetAddr      string
	redisNetAddr    string
	userTimeout     int
}

var (
	args Args
)

func init() {
	flag.StringVar(&args.addr, "addr", "localhost:2780", "address to serv.")
	flag.BoolVar(&args.testing, "testing", false, "whether to serv a testing page.")
	flag.BoolVar(&args.debug, "debug", false, "whether to enable debug tools.")
	flag.StringVar(&args.logicRPCNetAddr, "logic-rpc-net-addr", "tcp@localhost:6780", "logic rpc network address to listen.")
	flag.StringVar(&args.pprofAddr, "pprof-addr", "localhost:6060", "debug pprof http address.")
	flag.StringVar(&args.rpcNetAddr, "rpc-net-addr", "tcp@localhost:5780", "rpc network address to listen.")
	flag.StringVar(&args.redisNetAddr, "redis-net-addr", "tcp@localhost:6379", "redis network address.")
	flag.IntVar(&args.userTimeout, "user-timeout", 12, "user connection timeout.")
}
