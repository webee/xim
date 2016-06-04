package main

import (
	"flag"
	"xim/utils/argsutils"
)

// Args is app's arguments.
type Args struct {
	testing   bool
	debug     bool
	pprofAddr string

	pubAddrs *argsutils.StringSlice
	subAddrs *argsutils.StringSlice

	repAddrs *argsutils.StringSlice
	reqAddrs *argsutils.StringSlice
}

var (
	args = Args{
		pubAddrs: argsutils.NewStringSlice("tcp://:16783", "ipc:///tmp/xchat.proxy.pub.sock"),
		subAddrs: argsutils.NewStringSlice("tcp://:16784", "ipc:///tmp/xchat.proxy.sub.sock"),

		repAddrs: argsutils.NewStringSlice("tcp://:16787", "ipc:///tmp/xchat.proxy.rep.sock"),
		reqAddrs: argsutils.NewStringSlice("tcp://:16788", "ipc:///tmp/xchat.proxy.req.sock"),
	}
)

func init() {
	flag.BoolVar(&args.testing, "testing", false, "whether to serv a testing page.")
	flag.BoolVar(&args.debug, "debug", false, "whether to enable debug tools.")
	flag.StringVar(&args.pprofAddr, "pprof-addr", "localhost:6070", "debug pprof http address.")

	flag.Var(args.pubAddrs, "pub-addr", "proxy publish listen addresses.")
	flag.Var(args.subAddrs, "sub-addr", "proxy subscribe listen addresses.")

	flag.Var(args.repAddrs, "rep-addr", "proxy reply listen addresses.")
	flag.Var(args.reqAddrs, "req-addr", "proxy request listen addresses.")
}
