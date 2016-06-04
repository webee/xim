package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"
	"xim/utils/pprofutils"

	"gopkg.in/jcelliott/turnpike.v2"
	"strings"
	"time"
	"xim/apps/xchat/mid"
)

// XChatRouter is a wamp router for xchat.
type XChatRouter struct {
	*turnpike.WebsocketServer
}

const (
	latencyTopic = "latency"
	procedure    = "procedure"
)

var userkey = flag.String("userkey", "userkey", "app user key")
var debug = flag.Bool("debug", true, "debug mode")
var testing = flag.Bool("testing", true, "testing mode")
var endpoint = flag.String("endpoint", "/ws", "wamp router websocket url endpoint.")
var addr = flag.String("addr", "localhost:3699", "wamp server addr")
var pprofAddr = flag.String("pprofaddr", "0.0.0.0:3688", "pprof addr")
var topics map[int]string
var xchat *turnpike.Client

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()

	if *debug {
		pprofutils.StartPProfListen(*pprofAddr)
	}

	xchatRouter, err := NewXChatRouter([]byte(*userkey), *debug, *testing)
	if err != nil {
		log.Fatalln("create xchat channel failed.")
	}

	xchat, err = xchatRouter.GetLocalClient("xchat", nil)
	if err != nil {
		log.Fatalln("create xchat failed.", err)
	}

	Start(xchat)
	h, _, err := net.SplitHostPort(*addr)
	if err != nil {
		log.Fatalln("wrong addr", err)
	}

	port := 20000
	for i := 0; i < 1000; i++ {
		startRouter(xchatRouter, h+":"+strconv.Itoa(port+i))
	}

	setupSignal()
}

func startRouter(r *XChatRouter, addr string) {
	go func() {
		httpServeMux := http.NewServeMux()
		httpServeMux.Handle(*endpoint, r)
		httpServer := &http.Server{
			Handler: httpServeMux,
			Addr:    addr,
		}
		log.Println("http listen on: ", addr)
		log.Fatalln(httpServer.ListenAndServe())
	}()
}

func onJoin(args []interface{}, kwargs map[string]interface{}) {
	for _, v := range args {
		if id, ok := v.(int); ok {
			topics[id] = "topic:" + strconv.Itoa(id)
		}
	}
}

func onLatencyJoin(args []interface{}, kwargs map[string]interface{}) {
	for _, v := range args {
		details := v.(interface{})
		log.Println("Latenctyjoin: ", details)
	}
}

// Start starts the mid.
func Start(xchat *turnpike.Client) {
	if err := xchat.Subscribe(mid.URIWAMPSessionOnJoin, onJoin); err != nil {
		log.Fatalf("Error subscribing to %s: %s\n", mid.URIWAMPSessionOnJoin, err)
	}
	//
	//if err := xchat.Subscribe(latencyTopic, onLatencyJoin); err != nil {
	//	log.Fatalf("Error subscribing to %s: %s\n", mid.URIWAMPSessionOnJoin, err)
	//}

	if err := xchat.BasicRegister(procedure, MethodHandler); err != nil {
		log.Fatalf("Error basic register failed. %s", procedure)
	}
}

func MethodHandler(args []interface{}, kwargs map[string]interface{}) (result *turnpike.CallResult) {
	log.Println("rpc called")
	var err error

	for _, v := range args {
		if id, ok := v.(string); ok {
			log.Printf("....................publish topic{%s} %T\n", id, id)
			if strings.HasPrefix(id, latencyTopic) {
				err = xchat.Publish(latencyTopic, []interface{}{id}, nil)
			} else {
				err = xchat.Publish("topic:"+id, []interface{}{"replyFromServer: " + id}, nil)
			}
			if err != nil {
				log.Println("xchat publish failed.", err)
			}
		} else {
			log.Println("interface query failed.")
		}
	}
	return &turnpike.CallResult{Args: []interface{}{true}}
}

// setupSignal register signals handler and waiting for.
func setupSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	for {
		s := <-c
		log.Println("get a signal: ", s.String())
		switch s {
		case os.Interrupt, syscall.SIGTERM:
			return
		default:
			return
		}
	}
}

// NewXChatRouter creates a xchat router.
func NewXChatRouter(userKey []byte, debug, testing bool) (*XChatRouter, error) {
	if debug {
		turnpike.Debug()
	}
	realms := map[string]turnpike.Realm{
		"xchat": {},
	}
	if testing {
		realms["realm1"] = turnpike.Realm{}
	}

	s, err := turnpike.NewWebsocketServer(realms)
	if err != nil {
		return nil, err
	}

	// allow all origins.
	allowAllOrigin := func(r *http.Request) bool { return true }
	s.Upgrader.CheckOrigin = allowAllOrigin

	return &XChatRouter{
		WebsocketServer: s,
	}, nil
}

func sendMsg(c *turnpike.Client) {
	for {
		err := c.Publish(latencyTopic, []interface{}{1}, nil)
		if err != nil {
			log.Println("latency publish failed.", err)
		}

		time.Sleep(1 * time.Minute)
	}
}
