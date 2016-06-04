package main

import (
	"flag"
	"fmt"
	"log"
	"time"
	"xim/utils/argsutils"
	"xim/utils/nanorpc"
	"xim/xchat/logic/rpcservice/types"
)

var (
	addrs = argsutils.NewStringSlice("tcp://localhost:16787", "ipc:///tmp/xchat.logic.sock")
	txt   string
	count int
	print bool
)

func init() {
	flag.Var(addrs, "addr", "rpc service addresses")
	flag.StringVar(&txt, "txt", "TEST", "text to echo")
	flag.IntVar(&count, "c", 1, "count to call")
	flag.BoolVar(&print, "p", false, "print reply")
}

func main() {
	flag.Parse()

	client := nanorpc.NewClient(addrs.List())
	defer client.Close()
	t0 := time.Now()
	for i := 0; i < count; i++ {
		var reply string
		if err := client.Call(types.RPCXChatEcho, fmt.Sprintf("%s#%d", txt, i), &reply); err != nil {
			log.Fatalln("call echo error: ", err)
		}
		if print {
			log.Println("call echo: ", reply)
		}
	}
	d := time.Now().Sub(t0)
	log.Printf("%d calls: %d nano seconds.", count, d.Nanoseconds())
}
