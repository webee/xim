package main

import (
	"flag"
	"fmt"
	"log"
	"net/rpc"
	"time"
	"xim/xchat/logic/nanorpc"
	"xim/xchat/logic/rpcservice/types"

	"github.com/go-mangos/mangos"
	"github.com/go-mangos/mangos/protocol/req"
	"github.com/go-mangos/mangos/transport/ipc"
	"github.com/go-mangos/mangos/transport/tcp"
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

	s, err := req.NewSocket()
	if err != nil {
		log.Fatal("failed to open socket:", err)
	}

	s.SetOption(mangos.OptionRaw, true)
	s.AddTransport(tcp.NewTransport())
	s.AddTransport(ipc.NewTransport())
	// dial to load balancing rep/req proxy.
	if err := s.Dial(addr); err != nil {
		log.Fatal("can't dial on socket:", err)
	}

	client := rpc.NewClientWithCodec(nanorpc.NewNanoGobClientCodec(s))
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
