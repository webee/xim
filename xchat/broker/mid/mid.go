package mid

import (
	"log"
	"xim/utils/nanorpc"
	"xim/xchat/broker/logger"
	"xim/xchat/broker/router"
	"xim/xchat/logic/pub"

	ol "github.com/go-ozzo/ozzo-log"
	"gopkg.in/jcelliott/turnpike.v2"
)

var (
	l *ol.Logger
)

var (
	xchatLogic  *nanorpc.Client
	xchatSub    *pub.Subscriber
	xchat       *turnpike.Client
	emptyArgs   = []interface{}{}
	emptyKwargs = make(map[string]interface{})
)

func init() {
	l = logger.Logger.GetLogger("mid")
}

// Setup initialze mid.
func Setup(config *Config, xchatRouter *router.XChatRouter) {
	initXChatHTTPClient(config.Key, config.XChatHostURL)

	var err error
	xchatLogic = nanorpc.NewClient(config.LogicRPCAddr, config.RPCCallTimeout)

	xchat, err = xchatRouter.GetLocalClient("xchat", nil)
	if err != nil {
		log.Fatalf("create xchat error: %s", err)
	}

	if err := xchat.Subscribe(URIWAMPSessionOnJoin, sub(onJoin)); err != nil {
		log.Fatalf("Error subscribing to %s: %s", URIWAMPSessionOnJoin, err)
	}

	if err := xchat.Subscribe(URIWAMPSessionOnLeave, sub(onLeave)); err != nil {
		log.Fatalf("Error subscribing to %s: %s", URIWAMPSessionOnLeave, err)
	}

	if err := xchat.BasicRegister(URIXChatPing, call(ping)); err != nil {
		log.Fatalf("Error register %s: %s", URIXChatPing, err)
	}

	if err := xchat.BasicRegister(URIXChatSendMsg, call(sendMsg)); err != nil {
		log.Fatalf("Error register %s: %s", URIXChatSendMsg, err)
	}

	if err := xchat.Subscribe(URIXChatUserPub, sub(onPubMsg)); err != nil {
		log.Fatalf("Error subscribing to %s: %s", URIXChatUserPub, err)
	}

	if err := xchat.BasicRegister(URIXChatFetchChat, call(fetchChat)); err != nil {
		log.Fatalf("Error register %s: %s", URIXChatFetchChat, err)
	}

	if err := xchat.BasicRegister(URIXChatFetchChatMsgs, call(fetchChatMsgs)); err != nil {
		log.Fatalf("Error register %s: %s", URIXChatFetchChatMsgs, err)
	}

	if err := xchat.BasicRegister(URIXChatSyncChatRecv, call(syncChatRecv)); err != nil {
		log.Fatalf("Error register %s: %s", URIXChatSyncChatRecv, err)
	}

	if err := xchat.BasicRegister(URIXChatNewChat, call(newChat)); err != nil {
		log.Fatalf("Error register %s: %s", URIXChatNewChat, err)
	}

	if err := xchat.BasicRegister(URIXChatFetchChatList, call(fetchChatList)); err != nil {
		log.Fatalf("Error register %s: %s", URIXChatFetchChatList, err)
	}

	// Rooms
	if err := xchat.BasicRegister(URIXChatEnterRoom, call(enterRoom)); err != nil {
		log.Fatalf("Error register %s: %s", URIXChatEnterRoom, err)
	}

	if err := xchat.BasicRegister(URIXChatExitRoom, call(exitRoom)); err != nil {
		log.Fatalf("Error register %s: %s", URIXChatExitRoom, err)
	}

	// custome service
	if err := xchat.BasicRegister(URIXChatGetCsChat, call(getCsChat)); err != nil {
		log.Fatalf("Error register %s: %s", URIXChatGetCsChat, err)
	}

	xchatSub = pub.NewSubscriber(config.LogicPubAddr, 128)
	go handleMsg(xchatSub.Msgs())
}
