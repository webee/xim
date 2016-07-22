package main

import (
	"flag"
	"path"
	"time"
	"xim/utils/envutils"
)

// Args is app's arguments.
type Args struct {
	addr              string
	testing           bool
	debug             bool
	pprofAddr         string
	userKeyPath       string
	testUserKeyPath   string
	csUserKeyPath     string
	notifyUserKeyPath string
	logicRPCAddr      string
	rpcCallTimeout    time.Duration
}

var (
	args = Args{}
)

func init() {
	env := envutils.GetEnvDefault("XCHAT_ENV", "dev")
	flag.BoolVar(&args.testing, "testing", false, "whether to serv a testing page.")
	flag.BoolVar(&args.debug, "debug", false, "whether to enable debug tools.")
	flag.StringVar(&args.pprofAddr, "pprof-addr", "localhost:6062", "debug pprof http address.")
	flag.StringVar(&args.userKeyPath, "user-key-path", path.Join("conf", env, "xchat/user_key.txt"), "user key file path.")
	flag.StringVar(&args.testUserKeyPath, "test-user-key-path", path.Join("conf", env, "xchat/test_user_key.txt"), "test user key file path.")
	flag.StringVar(&args.csUserKeyPath, "cs-user-key-path", path.Join("conf", env, "xchat/cs_user_key.txt"), "custom service user key file path.")
	flag.StringVar(&args.notifyUserKeyPath, "notify-user-key-path", path.Join("conf", env, "xchat/notify_user_key.txt"), "notify user key file path.")
	flag.StringVar(&args.addr, "addr", "127.0.0.1:9981", "http api server listen addr.")
	flag.StringVar(&args.logicRPCAddr, "logic-rpc-addr", "tcp://:16787", "logic rpc addresses.")
	flag.DurationVar(&args.rpcCallTimeout, "rpc-timeout", 5*time.Second, "call rpc timeout.")
}
