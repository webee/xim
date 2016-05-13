package main

import "flag"

// Args is app's arguments.
type Args struct {
	addr             string
	testing          bool
	debug            bool
	pprofAddr        string
	dbDriverName     string
	dbDatasourceName string
	brokerURL        string
}

var (
	args Args
)

func init() {
	flag.StringVar(&args.addr, "addr", "localhost:6880", "address to serv.")
	flag.BoolVar(&args.testing, "testing", false, "whether to serv a testing page.")
	flag.BoolVar(&args.debug, "debug", false, "whether to enable debug tools.")
	flag.StringVar(&args.pprofAddr, "pprof-addr", "localhost:6070", "debug pprof http address.")
	flag.StringVar(&args.dbDriverName, "db-driver-name", "postgres", "database driver name.")
	flag.StringVar(&args.dbDatasourceName, "db-datasource-name", "postgres://xchat:xchat1234@localhost:5432/xim?sslmode=disable", "database datasoure name.")
	flag.StringVar(&args.brokerURL, "broker-url", "ws://127.0.0.1:48079/app-ws", "wamp broker url.")
}
