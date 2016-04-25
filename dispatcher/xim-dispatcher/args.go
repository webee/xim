package main

import "flag"

// Args is app's arguments.
type Args struct {
	rpcNetAddr    string
	testing       bool
	debug         bool
	pprofAddr     string
	redisServer   string
	redisPassword string
	redisNetAddr  string
}

var (
	args Args
)

func init() {
	flag.StringVar(&args.rpcNetAddr, "rpc-net-addr", "tcp@localhost:7780", "rpc network address to listen.")
	flag.BoolVar(&args.testing, "testing", false, "whether to serv a testing page.")
	flag.BoolVar(&args.debug, "debug", false, "whether to enable debug tools.")
	flag.StringVar(&args.pprofAddr, "pprof-addr", "localhost:6062", "debug pprof http address.")
	flag.StringVar(&args.redisServer, "redis-server", ":6379", "redis server.")
	flag.StringVar(&args.redisPassword, "redis-password", "", "redis password.")
	flag.StringVar(&args.redisNetAddr, "redis-net-addr", "tcp@localhost:6379", "redis network address.")
}
