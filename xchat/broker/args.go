package main

import (
	"flag"
	"path"
	"xim/utils/envutils"
)

// Args is app's arguments.
type Args struct {
	addr         string
	endpoint     string
	testWebDir   string
	testing      bool
	debug        bool
	brokerDebug  bool
	pprofAddr    string
	userKeyPath  string
	logicRPCAddr string
	logicPubAddr string
	XChatHostURL string
}

var (
	args = Args{}
)

func init() {
	env := envutils.GetEnvDefault("XCHAT_ENV", "dev")
	flag.BoolVar(&args.testing, "testing", false, "whether to serv a testing page.")
	flag.BoolVar(&args.debug, "debug", false, "whether to enable debug tools.")
	flag.BoolVar(&args.brokerDebug, "broker-debug", false, "whether to enable broker debug.")
	flag.StringVar(&args.pprofAddr, "pprof-addr", "localhost:6060", "debug pprof http address.")
	flag.StringVar(&args.userKeyPath, "user-key-path", path.Join("conf", env, "xchat/user_key.txt"), "user key file path.")
	flag.StringVar(&args.addr, "addr", "127.0.0.1:48080", "wamp router websocket listen addr.")
	flag.StringVar(&args.endpoint, "endpoint", "/ws", "wamp router websocket url endpoint.")
	flag.StringVar(&args.testWebDir, "test-web-dir", "xchat/broker/web", "test web dir.")
	flag.StringVar(&args.logicRPCAddr, "logic-rpc-addr", "tcp://:16787", "logic rpc addresses.")
	flag.StringVar(&args.logicPubAddr, "logic-pub-addr", "tcp://:16783", "logic pub address.")
	flag.StringVar(&args.XChatHostURL, "xchat-host-url", "http://localhost:9980", "xchat api host url.")
}
