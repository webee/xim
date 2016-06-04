package main

import "flag"

// Args is app's arguments.
type Args struct {
	testing          bool
	debug            bool
	dial             bool
	addr             string
	pprofAddr        string
	dbDriverName     string
	dbDatasourceName string
}

var (
	args Args
)

func init() {
	flag.BoolVar(&args.dial, "dial", false, "rpc service dial to addr(rep/req proxy)")
	flag.StringVar(&args.addr, "addr", "tcp://localhost:16787", "rpc listen/dial addr.")
	flag.BoolVar(&args.testing, "testing", false, "whether to serv a testing page.")
	flag.BoolVar(&args.debug, "debug", false, "whether to enable debug tools.")
	flag.StringVar(&args.pprofAddr, "pprof-addr", "localhost:6061", "debug pprof http address.")
	flag.StringVar(&args.dbDriverName, "db-driver-name", "postgres", "database driver name.")
	flag.StringVar(&args.dbDatasourceName, "db-datasource-name", "postgres://xchat:xchat1234@localhost:5432/xchat?sslmode=disable", "database datasoure name.")
}
