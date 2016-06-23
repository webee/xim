package main

import (
	"flag"
	"io/ioutil"
	"log"
	"path"
	"xim/utils/argsutils"
	"xim/utils/envutils"
)

// Args is app's arguments.
type Args struct {
	testing          bool
	debug            bool
	dial             bool
	addrs            *argsutils.StringSlice
	pubAddrs         *argsutils.StringSlice
	kafkaAddrs       *argsutils.StringSlice
	pprofAddr        string
	dbDriverName     string
	dbDatasourceName string
	dbMaxConn        int
	redisNetAddr     string
	redisPassword    string
	redisDB          int
}

var (
	args = Args{
		addrs:      argsutils.NewStringSlice("tcp://:16787", "ipc:///tmp/xchat.logic.rpc.sock"),
		pubAddrs:   argsutils.NewStringSlice("tcp://:16783", "ipc:///tmp/xchat.logic.pub.sock"),
		kafkaAddrs: argsutils.NewStringSlice("localhost:9092"),
	}
)

func readFile(path string) string {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalln("read file err:", err)
	}
	return string(data)
}

func init() {
	env := envutils.GetEnvDefault("XCHAT_ENV", "dev")
	flag.BoolVar(&args.dial, "dial", false, "rpc service dial to addr(rep/req proxy)")
	flag.Var(args.addrs, "addr", "rpc listen/dial addresses.")
	flag.Var(args.pubAddrs, "pub-addr", "publisher listen/dial addresses.")
	flag.Var(args.kafkaAddrs, "kafka-addr", "kafka broker addresses.")
	flag.BoolVar(&args.testing, "testing", false, "whether to serv a testing page.")
	flag.BoolVar(&args.debug, "debug", false, "whether to enable debug tools.")
	flag.StringVar(&args.pprofAddr, "pprof-addr", "localhost:6061", "debug pprof http address.")
	flag.StringVar(&args.dbDriverName, "db-driver-name", "postgres", "database driver name.")
	flag.StringVar(&args.dbDatasourceName, "db-datasource-name", readFile(path.Join("conf", env, "xchat/dbconn.txt")), "database datasoure name.")
	flag.IntVar(&args.dbMaxConn, "db-max-conn", 200, "database connection pool max connections.")
	flag.StringVar(&args.redisNetAddr, "redis-net-addr", "tcp@localhost:6379", "redis network address.")
	flag.StringVar(&args.redisPassword, "redis-password", "", "redis password.")
	flag.IntVar(&args.redisDB, "redis-db", 1, "redis db.")
}
