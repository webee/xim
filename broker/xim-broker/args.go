package main

import (
	"flag"
	"os"
	"path"
	"time"
)

// Args is app's arguments.
type Args struct {
	addr                 string
	testing              bool
	debug                bool
	logicRPCNetAddr      string
	pprofAddr            string
	rpcNetAddr           string
	redisNetAddr         string
	userTimeout          int
	httpReadTimeout      time.Duration
	httpWriteTimeout     time.Duration
	connHeartbeatTimeout time.Duration
	connWriteTimeout     time.Duration
	userKeyPath          string
}

var (
	args Args
)

func init() {
	env := os.Getenv("XIM_ENV")
	flag.StringVar(&args.addr, "addr", "localhost:2880", "address to serv.")
	flag.BoolVar(&args.testing, "testing", false, "whether to serv a testing page.")
	flag.BoolVar(&args.debug, "debug", false, "whether to enable debug tools.")
	flag.StringVar(&args.logicRPCNetAddr, "logic-rpc-net-addr", "tcp@localhost:6780", "logic rpc network address to listen.")
	flag.StringVar(&args.pprofAddr, "pprof-addr", "localhost:6060", "debug pprof http address.")
	flag.StringVar(&args.rpcNetAddr, "rpc-net-addr", "tcp@localhost:2780", "rpc network address to listen.")
	flag.StringVar(&args.redisNetAddr, "redis-net-addr", "tcp@localhost:6379", "redis network address.")
	flag.IntVar(&args.userTimeout, "user-timeout", 12, "user connection timeout(second).")
	flag.DurationVar(&args.httpReadTimeout, "http-read-timeout", 7*time.Second, "http read timeout.")
	flag.DurationVar(&args.httpWriteTimeout, "http-write-timeout", 7*time.Second, "http write timeout.")
	flag.DurationVar(&args.connHeartbeatTimeout, "conn-heartbeat-timeout", 12*time.Second, "connection heartbeat timeout.")
	flag.DurationVar(&args.connWriteTimeout, "conn-write-timeout", 7*time.Second, "connection write timeout.")
	flag.StringVar(&args.userKeyPath, "user-key-path", path.Join("conf", env, "xim/user_key.txt"), "user key file path.")
}
