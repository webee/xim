package main

import (
	"flag"
	"fmt"
	"log"
	"time"
	"xim/utils/nanorpc"
	"xim/xchat/logic/rpcservice/types"
)

var (
	addr  string
	txt   string
	count int
	print bool
)

func init() {
	flag.StringVar(&addr, "addr", "tcp://localhost:16780", "rpc service addr")
	flag.StringVar(&txt, "txt", "TEST", "text to echo")
	flag.IntVar(&count, "c", 1, "count to call")
	flag.BoolVar(&print, "p", false, "print reply")
}

func main() {
	flag.Parse()

	client := nanorpc.NewClient(addr)
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
