package main

import (
	"flag"
	"strings"
)

// Args is app's arguments.
type Args struct {
	testing          bool
	debug            bool
	pprofAddr        string
	dbDriverName     string
	dbDatasourceName string
	dbMaxConn        int
	kfkAddr          []string
	zkAddr           []string
	redisAddr        string
}

var (
	args = Args{}
)

func init() {
	flag.BoolVar(&args.testing, "testing", false, "whether to serv a testing page.")
	flag.BoolVar(&args.debug, "debug", false, "whether to enable debug tools.")
	flag.StringVar(&args.pprofAddr, "pprof-addr", "localhost:6061", "debug pprof http address.")
	flag.StringVar(&args.dbDriverName, "db-driver-name", "postgres", "database driver name.")
	flag.StringVar(&args.dbDatasourceName, "db-datasource-name", "postgres://xchat:xchat1234@localhost:5432/xchat?sslmode=disable", "database datasoure name.")
	flag.IntVar(&args.dbMaxConn, "db-max-conn", 200, "database connection pool max connections.")

	var tmpStr string
	flag.StringVar(&tmpStr, "kfk-addr", "localhost:9092", "the kafka addr")
	args.kfkAddr = strings.Split(tmpStr, ";")

	flag.StringVar(&tmpStr, "zk-addr", "localhost:2181", "the zookeeper addr")
	args.zkAddr = strings.Split(tmpStr, ";")

	flag.StringVar(&args.redisAddr, "redis-addr", "localhost:6379", "the redis addr")

}
