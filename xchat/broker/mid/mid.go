package mid

import (
	"log"
	"xim/xchat/broker/logger"
	"xim/xchat/broker/router"
	"xim/xchat/logic/service"

	ol "github.com/go-ozzo/ozzo-log"
	"github.com/valyala/gorpc"
	"gopkg.in/jcelliott/turnpike.v2"
)

var (
	l *ol.Logger
)

var (
	xchatDC     *gorpc.DispatcherClient
	xchat       *turnpike.Client
	emptyArgs   = []interface{}{}
	emptyKwargs = make(map[string]interface{})
)

// Init setup router.
func Init() {
	l = logger.Logger.GetLogger("mid")
}

// Setup initialze mid.
func Setup(config *Config, xchatRouter *router.XChatRouter) {
	c := gorpc.NewTCPClient(config.LogicRPCAddr)
	c.Start()
	d := service.NewServiceDispatcher()
	xchatDC = d.NewServiceClient(service.XChat.Name, c)

	xchat, err := xchatRouter.GetLocalClient("xchat", nil)
	if err != nil {
		log.Fatalf("create xchat error: %s", err)
	}

	if err := xchat.Subscribe(URIWAMPSessionOnJoin, onJoin); err != nil {
		log.Fatalf("Error subscribing to %s: %s", URIWAMPSessionOnJoin, err)
	}

	if err := xchat.Subscribe(URIWAMPSessionOnLeave, onLeave); err != nil {
		log.Fatalf("Error subscribing to %s: %s", URIWAMPSessionOnLeave, err)
	}

	if err := xchat.BasicRegister(URIXChatSendMsg, call(sendMsg)); err != nil {
		log.Fatalf("Error register %s: %s", URIXChatSendMsg, err)
	}

	if err := xchat.BasicRegister(URIXChatFetchChatMsgs, call(fetchChatMsg)); err != nil {
		log.Fatalf("Error register %s: %s", URIXChatFetchChatMsgs, err)
	}

	if err := xchat.BasicRegister(URIXChatNewChat, call(newChat)); err != nil {
		log.Fatalf("Error register %s: %s", URIXChatNewChat, err)
	}

	if err := xchat.BasicRegister(URIXChatFetchChatList, call(fetchChatList)); err != nil {
		log.Fatalf("Error register %s: %s", URIXChatNewChat, err)
	}
}
