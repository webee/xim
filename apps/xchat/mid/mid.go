package mid

import (
	"fmt"
	"log"
	"strings"
	"time"
	"xim/apps/xchat/db"
	"xim/apps/xchat/router"

	"gopkg.in/jcelliott/turnpike.v2"
)

// Mid is the mid of router and xim.
type Mid struct {
	xchat  *turnpike.Client
	xim    *XIMClient
	config *Config
}

var (
	msgID    int
	sessions map[uint64]bool
	mid      *Mid
)

// Setup initialze mid.
func Setup(config *Config, xchatRouter *router.XChatRouter) {
	xchat, err := xchatRouter.GetLocalClient("xchat", nil)
	if err != nil {
		log.Fatalln("create xchat error:", err)
	}

	xim := NewXIMClient(config)
	defer xim.Close()

	sessions = make(map[uint64]bool)
	mid = &Mid{
		xchat:  xchat,
		xim:    xim,
		config: config,
	}
	mid.Start()
}

// Start starts the mid.
func (m *Mid) Start() {
	xchat := m.xchat

	if m.config.Testing {
		if err := xchat.BasicRegister(URITestToUpper, call(toUpper)); err != nil {
			log.Fatalf("Error register %s: %s\n", URITestToUpper, err)
		}

		if err := xchat.BasicRegister(URITestAdd, call(add)); err != nil {
			log.Fatalf("Error register %s: %s\n", URITestAdd, err)
		}
	}

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
	delete(sessions, sessionID)
	log.Printf("<%d> left\n", sessionID)
	// unregister this user.
	m.xim.Unregister(sessionID)
}

// 处理用户注册
func (m *Mid) login(args []interface{}, kwargs map[string]interface{}) (result *turnpike.CallResult) {
	details := kwargs["details"].(map[string]interface{})
	sessionID := uint64(details["session"].(turnpike.ID))
	user := details["user"].(string)
	sessions[sessionID] = true
	if err := m.xim.Register(sessionID, user); err != nil {
		return &turnpike.CallResult{Err: turnpike.ErrInvalidArgument, Args: []interface{}{err}}
	}
	return &turnpike.CallResult{Args: []interface{}{true}}
}

// 用户发送消息
func (m *Mid) sendMsg(args []interface{}, kwargs map[string]interface{}) (result *turnpike.CallResult) {
	log.Printf("[rpc]%s: %v, %+v\n", URIXChatSendMsg, args, kwargs)
	details := kwargs["details"].(map[string]interface{})
	id := uint64(args[0].(float64))
	user := details["user"].(string)
	sessionID := uint64(details["session"].(turnpike.ID))

	chatID := uint64(args[1].(float64))
	msg := args[2]
	channel, err := db.GetChannelByChatIDAndUser(chatID, user)
	if err != nil {
		return &turnpike.CallResult{Err: turnpike.ErrInvalidArgument, Args: []interface{}{err}}
	}

	if err := m.xim.SendMsg(id, sessionID, channel, msg); err != nil {
		return &turnpike.CallResult{Err: turnpike.ErrInvalidArgument, Args: []interface{}{err}}
	}

	ts := time.Now().Unix()

	go func() {
		time.Sleep(5 * time.Millisecond)
		m.xchat.Publish(fmt.Sprintf(URIXChatUserReply, sessionID), nil, map[string]interface{}{
			"sn": id,
			"ok": true,
			"data": map[string]interface{}{
				"id": msgID,
				"ts": ts,
			},
		})
	}()
	go func() {
		msgID++
		time.Sleep(5 * time.Millisecond)
		for session := range sessions {
			if session != sessionID {
				m.xchat.Publish(fmt.Sprintf(URIXChatUserMsg, session), nil, map[string]interface{}{
					"chat_id": chatID,
					"user":    user,
					"id":      msgID,
					"ts":      ts,
					"msg":     msg,
				})
			}
		}
	}()
	return &turnpike.CallResult{Args: []interface{}{true}}
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

func toUpper(args []interface{}, kwargs map[string]interface{}) (result *turnpike.CallResult) {
	details := kwargs["details"].(map[string]interface{})
	user := details["user"].(string)
	sessionID := uint64(details["session"].(turnpike.ID))
	role := details["role"].(string)
	log.Printf("<%s:%s:%d> [rpc]%s: %v, %v, %v\n", role, user, sessionID, URITestToUpper, args, kwargs, details)
	s, ok := args[0].(string)
	if !ok {
		return &turnpike.CallResult{Err: turnpike.URI(URITestToUpper)}
	}
	res := strings.ToUpper(s)
	return &turnpike.CallResult{Args: []interface{}{res}}
}

func add(args []interface{}, kargs map[string]interface{}) (result *turnpike.CallResult) {
	details := kargs["details"].(map[string]interface{})
	user := details["user"].(string)
	sessionID := uint64(details["session"].(turnpike.ID))
	role := details["role"].(string)
	log.Printf("<%s:%s:%d> [rpc]%s: %v, %v, %v\n", role, user, sessionID, URITestAdd, args, kargs, details)
	a, ok1 := args[0].(float64)
	b, ok2 := args[1].(float64)
	if !(ok1 && ok2) {
		return &turnpike.CallResult{Err: turnpike.URI(URITestToUpper)}
	}
	res := a + b
	return &turnpike.CallResult{Args: []interface{}{res}}
}
