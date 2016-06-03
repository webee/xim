package main

import "flag"

// Args is app's arguments.
type Args struct {
	testing   bool
	debug     bool
	pprofAddr string
	repAddr   string
	reqAddr   string
	pubAddr   string
	subAddr   string
}

var (
	args Args
)

func init() {
	flag.BoolVar(&args.testing, "testing", false, "whether to serv a testing page.")
	flag.BoolVar(&args.debug, "debug", false, "whether to enable debug tools.")
	flag.StringVar(&args.pprofAddr, "pprof-addr", "localhost:6070", "debug pprof http address.")
	flag.StringVar(&args.pubAddr, "pub-addr", "tcp://:16783", "proxy publish listen addr.")
	flag.StringVar(&args.subAddr, "sub-addr", "tcp://:16784", "proxy subscribe listen addr.")
	flag.StringVar(&args.repAddr, "rep-addr", "tcp://:16787", "proxy reply listen addr.")
	flag.StringVar(&args.reqAddr, "req-addr", "tcp://:16788", "proxy request listen addr.")
}
