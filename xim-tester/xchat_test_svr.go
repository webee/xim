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
)

// XChatRouter is a wamp router for xchat.
type XChatRouter struct {
	*turnpike.WebsocketServer
}

var userkey = flag.String("userkey", "userkey", "app user key")
var debug = flag.Bool("debug", true, "debug mode")
var testing = flag.Bool("testing", true, "testing mode")
var endpoint = flag.String("endpoint", "/ws", "wamp router websocket url endpoint.")
var addr = flag.String("addr", "localhost:3699", "wamp server addr")
var pprofAddr = flag.String("pprofaddr", "0.0.0.0:3688", "pprof addr")

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

	_, err = xchatRouter.GetLocalClient("xchat", nil)
	if err != nil {
		log.Fatalln("create xchat failed.", err)
	}

	//	Start(xchat)
	h, _, err := net.SplitHostPort(*addr)
	if err != nil {
		log.Fatalln("wrong addr", err)
	}
	//	port, err := strconv.Atoi(p)
	//	if err != nil {
	//		log.Fatalln("wrong addr", err)
	//	}
	port := 20000
	for i := 0; i < 1000; i++ {
		startRouter(xchatRouter, h+":"+strconv.Itoa(port+i))
	}

	setupSignal()
}

func startRouter(r *XChatRouter, addr string) {
	go func() {
		httpServeMux := http.NewServeMux()
		//		if testing {
		//			httpServeMux.Handle("/", http.FileServer(http.Dir(args.testWebDir)))
		//		}
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
		details := v.(interface{})
		log.Println("join: ", details)
	}
}

func onLeave(args []interface{}, kwargs map[string]interface{}) {
	sessionID := uint64(args[0].(turnpike.ID))
	//	m.xim.Unregister(sessionID)
	log.Printf("<%d> left\n", sessionID)
}

// Start starts the mid.
func Start(xchat *turnpike.Client) {
	//	if err := xchat.Subscribe(mid.URIWAMPSessionOnJoin, onJoin); err != nil {
	//		log.Fatalf("Error subscribing to %s: %s\n", mid.URIWAMPSessionOnJoin, err)
	//	}

	//	if err := xchat.Subscribe(mid.URIWAMPSessionOnLeave, onLeave); err != nil {
	//		log.Fatalf("Error subscribing to %s: %s\n", mid.URIWAMPSessionOnLeave, err)
	//	}

	//	if err := xchat.BasicRegister(URIXChatLogin, call(m.login)); err != nil {
	//		log.Fatalf("Error register %s: %s\n", URIXChatLogin, err)
	//	}

	//	if err := xchat.BasicRegister(URIXChatSendMsg, call(m.sendMsg)); err != nil {
	//		log.Fatalf("Error register %s: %s\n", URIXChatSendMsg, err)
	//	}

	//	if err := xchat.BasicRegister(URIXChatFetchChatMsgs, call(m.fetchChatMsg)); err != nil {
	//		log.Fatalf("Error register %s: %s\n", URIXChatFetchChatMsgs, err)
	//	}

	//	if err := xchat.BasicRegister(URIXChatNewChat, call(m.newChat)); err != nil {
	//		log.Fatalf("Error register %s: %s\n", URIXChatNewChat, err)
	//	}

	//	if err := xchat.BasicRegister(URIXChatFetchChatList, call(m.fetchChatList)); err != nil {
	//		log.Fatalf("Error register %s: %s\n", URIXChatNewChat, err)
	//	}
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
		"xchat": {
		//				Authorizer:  NewUserRoleAuthorizer(nil),
		//				Interceptor: NewDetailsInterceptor(roleIsUser, nil, "details"),
		//				CRAuthenticators: map[string]turnpike.CRAuthenticator{
		//					"jwt": &jwtAuth{key: userKey},
		//				},
		},
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
