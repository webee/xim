package main

import (
	"flag"
	"path"
	"xim/utils/argsutils"
	"xim/utils/envutils"
)

// Args is app's arguments.
type Args struct {
	addr          string
	endpoint      string
	testWebDir    string
	testing       bool
	debug         bool
	pprofAddr     string
	userKeyPath   string
	logicRPCAddrs *argsutils.StringSlice
}

var (
	args = Args{
		logicRPCAddrs: argsutils.NewStringSlice("tcp://localhost:16787", "ipc:///tmp/xchat.logic.sock"),
	}
)

func init() {
	env := envutils.GetEnvDefault("XCHAT_ENV", "dev")
	flag.BoolVar(&args.testing, "testing", false, "whether to serv a testing page.")
	flag.BoolVar(&args.debug, "debug", false, "whether to enable debug tools.")
	flag.StringVar(&args.pprofAddr, "pprof-addr", "localhost:6060", "debug pprof http address.")
	flag.StringVar(&args.userKeyPath, "user-key-path", path.Join("conf", env, "xchat/user_key.txt"), "user key file path.")
	flag.StringVar(&args.addr, "addr", "127.0.0.1:48080", "wamp router websocket listen addr.")
	flag.StringVar(&args.endpoint, "endpoint", "/ws", "wamp router websocket url endpoint.")
	flag.StringVar(&args.testWebDir, "test-web-dir", "xchat/broker/web", "test web dir.")
	flag.Var(args.logicRPCAddrs, "logic-rpc-addr", "logic rpc addresses.")
}
