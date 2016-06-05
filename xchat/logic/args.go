package main

import (
	"flag"
	"xim/utils/argsutils"
)

// Args is app's arguments.
type Args struct {
	testing          bool
	debug            bool
	dial             bool
	addrs            *argsutils.StringSlice
	pubAddrs         *argsutils.StringSlice
	pprofAddr        string
	dbDriverName     string
	dbDatasourceName string
}

var (
	args = Args{
		addrs:    argsutils.NewStringSlice("tcp://:16787", "ipc:///tmp/xchat.logic.rpc.sock"),
		pubAddrs: argsutils.NewStringSlice("tcp://:16783", "ipc:///tmp/xchat.logic.pub.sock"),
	}
)

func init() {
	flag.BoolVar(&args.dial, "dial", false, "rpc service dial to addr(rep/req proxy)")
	flag.Var(args.addrs, "addr", "rpc listen/dial addresses.")
	flag.Var(args.pubAddrs, "pub-addr", "publisher listen/dial addresses.")
	flag.BoolVar(&args.testing, "testing", false, "whether to serv a testing page.")
	flag.BoolVar(&args.debug, "debug", false, "whether to enable debug tools.")
	flag.StringVar(&args.pprofAddr, "pprof-addr", "localhost:6061", "debug pprof http address.")
	flag.StringVar(&args.dbDriverName, "db-driver-name", "postgres", "database driver name.")
	flag.StringVar(&args.dbDatasourceName, "db-datasource-name", "postgres://xchat:xchat1234@localhost:5432/xchat?sslmode=disable", "database datasoure name.")
}
