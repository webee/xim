package mid

import (
	"log"
	"math"
	"math/rand"
	"time"
	"xim/utils/nanorpc"
	"xim/xchat/broker/logger"
	"xim/xchat/broker/router"
	"xim/xchat/logic/pub"
	"xim/xchat/xchat-http-client"

	ol "github.com/go-ozzo/ozzo-log"
	"gopkg.in/webee/turnpike.v2"
)

var (
	l *ol.Logger
)

var (
	instanceID      uint64
	xchatHTTPClient *xchathttpclient.XChatHTTPClient
	xchatLogic      *nanorpc.Client
	xchatSub        *pub.Subscriber
	xchat           *turnpike.Client
	realm           *turnpike.Realm
	emptyArgs       = []interface{}{}
	emptyKwargs     = make(map[string]interface{})
)

func init() {
	l = logger.Logger.GetLogger("mid")
	rand.Seed(time.Now().UnixNano())
}

// Setup initialze mid.
func Setup(config *Config, xchatRouter *router.XChatRouter) {
	instanceID = uint64(rand.Int63n(math.MaxInt64))

	xchatHTTPClient = xchathttpclient.NewXChatHTTPClient(config.Key, config.XChatHostURL)

	var (
		err error
		ok  bool
	)
	xchatLogic = nanorpc.NewClient(config.LogicRPCAddr, config.RPCCallTimeout)

	realm, ok = xchatRouter.GetRealm("xchat")
	if !ok {
		log.Fatalf("realm xchat not exists")
	}

	xchat, err = xchatRouter.GetLocalClientWithSize(100, "xchat", nil)
	if err != nil {
		log.Fatalf("create xchat error: %s", err)
	}

	subscribeTopic(xchat, URIWAMPSessionOnJoin, onJoin)
	subscribeTopic(xchat, URIWAMPSessionOnLeave, onLeave)

	// ping: test rpc.
	registerSessionProcedure(xchat, URIXChatPing, ping)

	registerSessionProcedure(xchat, URIXChatPubUserInfo, pubUserInfo)
	subscribeSessionTopic(xchat, URIXChatPubUserInfo, onPubUserInfo)

	registerSessionProcedure(xchat, URIXChatPubUserStatusInfo, pubUserStatusInfo)
	subscribeSessionTopic(xchat, URIXChatPubUserStatusInfo, onPubUserStatusInfo)

	registerSessionProcedure(xchat, URIXChatSendMsg, sendMsg)
	registerSessionProcedure(xchat, URIXChatSendNotify, sendNotify)
	subscribeSessionTopic(xchat, URIXChatPubNotify, onPubNotify)
	registerSessionProcedure(xchat, URIXChatSendUserNotify, sendUserNotify)
	subscribeSessionTopic(xchat, URIXChatPubUserNotify, onPubUserNotify)

	registerSessionProcedure(xchat, URIXChatJoinChat, joinChat)
	registerSessionProcedure(xchat, URIXChatExitChat, exitChat)

	registerSessionProcedure(xchat, URIXChatSetChatTitle, setChatTitle)

	registerSessionProcedure(xchat, URIXChatFetchChat, fetchChat)
	registerSessionProcedure(xchat, URIXChatFetchChatMembers, fetchChatMembers)
	registerSessionProcedure(xchat, URIXChatFetchChatMsgs, fetchChatMsgs)

	registerSessionProcedure(xchat, URIXChatSetChat, setChat)
	registerSessionProcedure(xchat, URIXChatSyncChatRecv, syncChatRecv)
	registerSessionProcedure(xchat, URIXChatNewChat, newChat)
	registerSessionProcedure(xchat, URIXChatFetchChatList, fetchChatList)

	// Rooms
	registerSessionProcedure(xchat, URIXChatEnterRoom, enterRoom)
	registerSessionProcedure(xchat, URIXChatExitRoom, exitRoom)

	// custome service
	registerSessionProcedure(xchat, URIXChatGetCsChat, getCsChat)

	xchatSub = pub.NewSubscriber(config.LogicPubAddr, 128)
	go handleMsg(xchatSub.Msgs())

	// tasks
	go TaskUpdatingOnlineUsers()
}

func registerProcedure(client *turnpike.Client, uri string, p Procedure) {
	if err := p.registerTo(client, uri); err != nil {
		log.Fatalf("Error register %s: %s", uri, err)
	}
}

func registerSessionProcedure(client *turnpike.Client, uri string, sp SessionProcedure) {
	if err := sp.registerTo(client, uri); err != nil {
		log.Fatalf("Error register %s: %s", uri, err)
	}
}

func subscribeTopic(client *turnpike.Client, topic string, s Subscriber) {
	if err := s.subscribeTo(client, topic); err != nil {
		log.Fatalf("Error subscribing to %s: %s", topic, err)
	}
}

func subscribeSessionTopic(client *turnpike.Client, topic string, ss SessionSubscriber) {
	if err := ss.subscribeTo(client, topic); err != nil {
		log.Fatalf("Error subscribing to %s: %s", topic, err)
	}
}
