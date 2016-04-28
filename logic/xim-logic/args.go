package main

import (
	"flag"
	"os"
	"path"
)

// Args is app's arguments.
type Args struct {
	addr                 string
	rpcNetAddr           string
	dispatcherRPCNetAddr string
	testing              bool
	debug                bool
	pprofAddr            string
	saltPath             string
	appKeyPath           string
	userKeyPath          string
}

var (
	args Args
)

func init() {
	env := os.Getenv("XIM_ENV")
	flag.StringVar(&args.addr, "addr", "localhost:6880", "address to serv.")
	flag.StringVar(&args.rpcNetAddr, "rpc-net-addr", "tcp@localhost:6780", "rpc network address to listen.")
	flag.StringVar(&args.dispatcherRPCNetAddr, "dispatcher-rpc-net-addr", "tcp@localhost:7780", "dispatcher rpc network address to listen.")
	flag.BoolVar(&args.testing, "testing", false, "whether to serv a testing page.")
	flag.BoolVar(&args.debug, "debug", false, "whether to enable debug tools.")
	flag.StringVar(&args.pprofAddr, "pprof-addr", "localhost:6061", "debug pprof http address.")
	flag.StringVar(&args.saltPath, "salt-path", path.Join("conf", env, "xim/salt.txt"), "app key file path.")
	flag.StringVar(&args.appKeyPath, "app-key-path", path.Join("conf", env, "xim/app_key.txt"), "app key file path.")
	flag.StringVar(&args.userKeyPath, "user-key-path", path.Join("conf", env, "xim/user_key.txt"), "user key file path.")
}
