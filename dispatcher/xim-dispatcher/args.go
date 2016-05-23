package main

import "flag"

// Args is app's arguments.
type Args struct {
	rpcNetAddr       string
	testing          bool
	debug            bool
	pprofAddr        string
	redisNetAddr     string
	redisPassword    string
	dbDriverName     string
	dbDatasourceName string
	mangoURL         string
}

var (
	args Args
)

func init() {
	flag.StringVar(&args.rpcNetAddr, "rpc-net-addr", "tcp@localhost:7780", "rpc network address to listen.")
	flag.BoolVar(&args.testing, "testing", false, "whether to serv a testing page.")
	flag.BoolVar(&args.debug, "debug", false, "whether to enable debug tools.")
	flag.StringVar(&args.pprofAddr, "pprof-addr", "localhost:6062", "debug pprof http address.")
	flag.StringVar(&args.redisNetAddr, "redis-net-addr", "tcp@localhost:6379", "redis network address.")
	flag.StringVar(&args.redisPassword, "redis-password", "", "redis password.")
	flag.StringVar(&args.dbDriverName, "db-driver-name", "postgres", "database driver name.")
	flag.StringVar(&args.dbDatasourceName, "db-datasource-name", "postgres://xim:xim1234@localhost:5432/xim?sslmode=disable", "database datasoure name.")
	flag.StringVar(&args.mangoURL, "mango-url", "localhost:27017", "mango db url.")
}
