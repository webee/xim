package main

import (
	"flag"
	"path"
	"xim/utils/envutils"
)

// Args is app's arguments.
type Args struct {
	addr             string
	endpoint         string
	testWebDir       string
	testing          bool
	debug            bool
	pprofAddr        string
	dbDriverName     string
	dbDatasourceName string
	ximHostURL       string
	ximApp           string
	ximPassword      string
	ximAppWsURL      string
	userKeyPath      string
}

var (
	args Args
)

func init() {
	env := envutils.GetEnvDefault("XCHAT_ENV", "dev")
	flag.BoolVar(&args.testing, "testing", false, "whether to serv a testing page.")
	flag.BoolVar(&args.debug, "debug", false, "whether to enable debug tools.")
	flag.StringVar(&args.pprofAddr, "pprof-addr", "localhost:6070", "debug pprof http address.")
	flag.StringVar(&args.dbDriverName, "db-driver-name", "postgres", "database driver name.")
	flag.StringVar(&args.dbDatasourceName, "db-datasource-name", "postgres://xchat:xchat1234@localhost:5432/xim?sslmode=disable", "database datasoure name.")
	flag.StringVar(&args.ximHostURL, "xim-host-url", "http://localhost:6980", "xim api host url.")
	flag.StringVar(&args.ximApp, "xim-app", "test", "xim app name.")
	flag.StringVar(&args.ximPassword, "xim-password", "test1234", "xim password.")
	flag.StringVar(&args.ximAppWsURL, "xim-app-ws-url", "ws://127.0.0.1:2980/ws", "xim app websocket url.")
	flag.StringVar(&args.userKeyPath, "user-key-path", path.Join("conf", env, "xchat/user_key.txt"), "user key file path.")
	flag.StringVar(&args.addr, "addr", "127.0.0.1:48080", "wamp router websocket listen addr.")
	flag.StringVar(&args.endpoint, "endpoint", "/ws", "wamp router websocket url endpoint.")
	flag.StringVar(&args.testWebDir, "test-web-dir", "apps/xchat/xchat-app/web", "test web dir.")
}
