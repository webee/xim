package mid

import (
	"fmt"
	"log"
	"xim/apps/xchat/db"
	"xim/apps/xchat/router"
	"xim/broker/proto"
	"xim/utils/msgutils"

	"gopkg.in/jcelliott/turnpike.v2"
)

// Mid is the mid of router and xim.
type Mid struct {
	xchat  *turnpike.Client
	xim    *XIMClient
	config *Config
}

var (
	mid *Mid
)

// Setup initialze mid.
func Setup(config *Config, xchatRouter *router.XChatRouter) {
	initXimHTTPClient(config.XIMApp, config.XIMPassword, config.XIMHostURL)

	xchat, err := xchatRouter.GetLocalClient("xchat", nil)
	if err != nil {
		log.Fatalln("create xchat error:", err)
	}

	mid = &Mid{
		xchat:  xchat,
		config: config,
	}
	mid.xim = NewXIMClient(config, mid.handleMsg)

	mid.Start()
}

// Start starts the mid.
func (m *Mid) Start() {
	xchat := m.xchat

	if err := xchat.Subscribe(URIWAMPSessionOnJoin, m.onJoin); err != nil {
		log.Fatalf("Error subscribing to %s: %s\n", URIWAMPSessionOnJoin, err)
	}

	if err := xchat.Subscribe(URIWAMPSessionOnLeave, m.onLeave); err != nil {
		log.Fatalf("Error subscribing to %s: %s\n", URIWAMPSessionOnLeave, err)
	}

	if err := xchat.BasicRegister(URIXChatLogin, call(m.login)); err != nil {
		log.Fatalf("Error register %s: %s\n", URIXChatLogin, err)
	}

	if err := xchat.BasicRegister(URIXChatSendMsg, call(m.sendMsg)); err != nil {
		log.Fatalf("Error register %s: %s\n", URIXChatSendMsg, err)
	}
}

// 处理用户连接
func (m *Mid) onJoin(args []interface{}, kwargs map[string]interface{}) {
	details := args[0].(map[string]interface{})
	log.Println("join: ", details)
}

// 处理用户断开注销
func (m *Mid) onLeave(args []interface{}, kwargs map[string]interface{}) {
	sessionID := uint64(args[0].(turnpike.ID))
	m.xim.Unregister(sessionID)
	log.Printf("<%d> left\n", sessionID)
}

// 处理用户注册
func (m *Mid) login(args []interface{}, kwargs map[string]interface{}) (result *turnpike.CallResult) {
	details := kwargs["details"].(map[string]interface{})
	sessionID := uint64(details["session"].(turnpike.ID))
	user := details["user"].(string)
	log.Println("login:", sessionID)
	if err := m.xim.Register(sessionID, user); err != nil {
		return &turnpike.CallResult{Err: turnpike.ErrInvalidArgument, Args: []interface{}{err}}
	}
	return &turnpike.CallResult{Args: []interface{}{true}}
}

// 用户发送消息
func (m *Mid) sendMsg(args []interface{}, kwargs map[string]interface{}) (result *turnpike.CallResult) {
	log.Printf("[rpc]%s: %v, %+v\n", URIXChatSendMsg, args, kwargs)
	details := kwargs["details"].(map[string]interface{})
	user := details["user"].(string)
	sessionID := uint64(details["session"].(turnpike.ID))

	chatID := uint64(args[0].(float64))
	msg := args[1]
	channel, err := db.GetChannelByChatIDAndUser(chatID, user)
	if err != nil {
		return &turnpike.CallResult{Err: turnpike.ErrInvalidArgument, Args: []interface{}{err}}
	}

	id, ts, err := m.xim.SendMsg(sessionID, channel, msg)
	if err != nil {
		return &turnpike.CallResult{Err: turnpike.ErrInvalidArgument, Args: []interface{}{false, 1, err.Error()}}
	}

	return &turnpike.CallResult{Args: []interface{}{true, id, ts}}
}

func (m *Mid) handleMsg(msg msgutils.Message) {
	switch x := msg.(type) {
	case *proto.Push:
		sessionID, err := m.xim.getSessionIDbyUID(x.UID)
		if err != nil {
			return
		}

		chatID := 1
		_ = m.xchat.Publish(fmt.Sprintf(URIXChatUserMsg, sessionID), nil, map[string]interface{}{
			"chat_id": chatID,
			"user":    x.User,
			"id":      x.ID,
			"ts":      x.Ts,
			"msg":     x.Msg,
		})
	}
}

func call(handler turnpike.BasicMethodHandler) turnpike.BasicMethodHandler {
	return func(args []interface{}, kargs map[string]interface{}) (result *turnpike.CallResult) {
		defer func() {
			if r := recover(); r != nil {
				result = &turnpike.CallResult{Err: turnpike.ErrInvalidArgument, Args: []interface{}{r}}
			}
		}()
		return handler(args, kargs)
	}
}
