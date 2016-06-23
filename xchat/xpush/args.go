package main

import (
	"flag"
	"log"
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
	zkAddr           string
	redisAddr        string
	xgtest           bool
	pushInterval     int64
	apiLogHost       string
	userInfoHost     string
}

var (
	args = Args{}
)

func init() {
	flag.BoolVar(&args.testing, "testing", false, "whether to serv a testing page.")
	flag.BoolVar(&args.debug, "debug", false, "whether to enable debug tools.")
	flag.StringVar(&args.pprofAddr, "pprof-addr", "localhost:6061", "debug pprof http address.")

	var tmpStr string
	flag.StringVar(&tmpStr, "kfk-addr", "localhost:9092", "the kafka addr")
	args.kfkAddr = strings.Split(tmpStr, ";")
	log.Println(args.kfkAddr)

	flag.StringVar(&args.zkAddr, "zk-addr", "localhost:2181/kafka", "the zookeeper addr")
	flag.StringVar(&args.redisAddr, "redis-addr", "localhost:6379", "the redis addr")
	flag.BoolVar(&args.xgtest, "xgtest", true, "is xinge test environment")
	flag.Int64Var(&args.pushInterval, "push-interval", 60, "push offline msg interval")
	flag.StringVar(&args.apiLogHost, "apilog-host", "http://apilogdoc.engdd.com", "api log host")
	flag.StringVar(&args.userInfoHost, "user-info-host", "http://test.engdd.com", "user info host")
}
