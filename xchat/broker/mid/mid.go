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

	ol "github.com/go-ozzo/ozzo-log"
	"gopkg.in/webee/turnpike.v2"
)

var (
	l *ol.Logger
)

var (
	instanceID  uint64
	xchatLogic  *nanorpc.Client
	xchatSub    *pub.Subscriber
	xchat       *turnpike.Client
	emptyArgs   = []interface{}{}
	emptyKwargs = make(map[string]interface{})
)

func init() {
	l = logger.Logger.GetLogger("mid")
	rand.Seed(time.Now().UnixNano())
}

// Setup initialze mid.
func Setup(config *Config, xchatRouter *router.XChatRouter) {
	instanceID = uint64(rand.Int63n(math.MaxInt64))

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

	if err := xchat.BasicRegister(URIXChatPubUserInfo, call(pubUserInfo)); err != nil {
		log.Fatalf("Error register %s: %s", URIXChatPubUserInfo, err)
	}

	if err := xchat.Subscribe(URIXChatPubUserInfo, sub(onPubUserInfo)); err != nil {
		log.Fatalf("Error subscribing to %s: %s", URIXChatPubUserInfo, err)
	}

	if err := xchat.BasicRegister(URIXChatSendMsg, call(sendMsg)); err != nil {
		log.Fatalf("Error register %s: %s", URIXChatSendMsg, err)
	}

	if err := xchat.Subscribe(URIXChatPubMsg, sub(onPubMsg)); err != nil {
		log.Fatalf("Error subscribing to %s: %s", URIXChatPubMsg, err)
	}

	// 会话
	if err := xchat.BasicRegister(URIXChatJoinChat, call(joinChat)); err != nil {
		log.Fatalf("Error register %s: %s", URIXChatJoinChat, err)
	}

	if err := xchat.BasicRegister(URIXChatExitChat, call(exitChat)); err != nil {
		log.Fatalf("Error register %s: %s", URIXChatExitChat, err)
	}

	if err := xchat.BasicRegister(URIXChatFetchChat, call(fetchChat)); err != nil {
		log.Fatalf("Error register %s: %s", URIXChatFetchChat, err)
	}

	if err := xchat.BasicRegister(URIXChatFetchChatMembers, call(fetchChatMembers)); err != nil {
		log.Fatalf("Error register %s: %s", URIXChatFetchChatMembers, err)
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

	// tasks
	go TaskUpdatingOnlineUsers()
}
