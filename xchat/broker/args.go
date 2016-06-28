package main

import (
	"flag"
	"path"
	"time"
	"xim/utils/envutils"
)

// Args is app's arguments.
type Args struct {
	addr            string
	endpoint        string
	testWebDir      string
	testing         bool
	debug           bool
	brokerDebug     bool
	writeTimeout    time.Duration
	idleTimeout     time.Duration
	pprofAddr       string
	userKeyPath     string
	testUserKeyPath string
	csUserKeyPath   string
	logicRPCAddr    string
	logicPubAddr    string
	xchatHostURL    string
	rpcCallTimeout  time.Duration
}

var (
	args = Args{}
)

func init() {
	env := envutils.GetEnvDefault("XCHAT_ENV", "dev")
	flag.BoolVar(&args.testing, "testing", false, "whether to serv a testing page.")
	flag.BoolVar(&args.debug, "debug", false, "whether to enable debug tools.")
	flag.BoolVar(&args.brokerDebug, "broker-debug", false, "whether to enable broker debug.")
	flag.DurationVar(&args.writeTimeout, "write-timeout", 20*time.Second, "router write timeout.")
	flag.DurationVar(&args.idleTimeout, "idle-timeout", 10*time.Minute, "client idle timeout.")
	flag.StringVar(&args.pprofAddr, "pprof-addr", "localhost:6060", "debug pprof http address.")
	flag.StringVar(&args.userKeyPath, "user-key-path", path.Join("conf", env, "xchat/user_key.txt"), "user key file path.")
	flag.StringVar(&args.testUserKeyPath, "test-user-key-path", path.Join("conf", env, "xchat/test_user_key.txt"), "test user key file path.")
	flag.StringVar(&args.csUserKeyPath, "cs-user-key-path", path.Join("conf", env, "xchat/cs_user_key.txt"), "custom service user key file path.")
	flag.StringVar(&args.addr, "addr", "127.0.0.1:48080", "wamp router websocket listen addr.")
	flag.StringVar(&args.endpoint, "endpoint", "/ws", "wamp router websocket url endpoint.")
	flag.StringVar(&args.testWebDir, "test-web-dir", "xchat/broker/web", "test web dir.")
	flag.StringVar(&args.logicRPCAddr, "logic-rpc-addr", "tcp://:16787", "logic rpc addresses.")
	flag.StringVar(&args.logicPubAddr, "logic-pub-addr", "tcp://:16783", "logic pub address.")
	flag.StringVar(&args.xchatHostURL, "xchat-host-url", "http://localhost:9980", "xchat api host url.")
	flag.DurationVar(&args.rpcCallTimeout, "rpc-timeout", 5*time.Second, "call rpc timeout.")
}
