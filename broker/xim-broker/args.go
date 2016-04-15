package main

import "flag"

// Args is app's arguments.
type Args struct {
	addr    string
	testing bool
	debug   bool
}

var (
	args Args
)

func init() {
	flag.StringVar(&args.addr, "addr", "localhost:2780", "address to serv.")
	flag.BoolVar(&args.testing, "testing", false, "whether to serv a testing page.")
	flag.BoolVar(&args.debug, "debug", false, "whether to enable debug tools.")
}
