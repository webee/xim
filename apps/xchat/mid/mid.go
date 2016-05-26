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
	mid         *Mid
	emptyArgs   = []interface{}{}
	emptyKwargs = make(map[string]interface{})
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

	if err := xchat.BasicRegister(URIXChatFetchChatMsgs, call(m.fetchChatMsg)); err != nil {
		log.Fatalf("Error register %s: %s\n", URIXChatFetchChatMsgs, err)
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

func getSessionFromDetails(d interface{}) (sessionID uint64, user string) {
	details := d.(map[string]interface{})
	sessionID = uint64(details["session"].(turnpike.ID))
	user = details["user"].(string)
	return
}

// 处理用户注册
func (m *Mid) login(args []interface{}, kwargs map[string]interface{}) (result *turnpike.CallResult) {
	sessionID, user := getSessionFromDetails(kwargs["details"])
	log.Println("login:", sessionID)
	if err := m.xim.Register(sessionID, user); err != nil {
		return &turnpike.CallResult{Args: []interface{}{false, 1, err.Error()}}
	}
	return &turnpike.CallResult{Args: []interface{}{true}}
}

// 用户发送消息
func (m *Mid) sendMsg(args []interface{}, kwargs map[string]interface{}) (result *turnpike.CallResult) {
	log.Printf("[rpc]%s: %v, %+v\n", URIXChatSendMsg, args, kwargs)
	sessionID, user := getSessionFromDetails(kwargs["details"])

	chatID := uint64(args[0].(float64))
	channel, err := db.GetChannelByChatIDAndUser(chatID, user)
	if err != nil {
		return &turnpike.CallResult{Args: []interface{}{false, 1, err.Error()}}
	}

	msg := args[1]
	id, ts, err := m.xim.SendMsg(sessionID, channel, msg)
	if err != nil {
		return &turnpike.CallResult{Args: []interface{}{false, 1, err.Error()}}
	}

	return &turnpike.CallResult{Args: []interface{}{true, id, ts}}
}

// 获取会话列表
func (m *Mid) fetchChatList(args []interface{}, kwargs map[string]interface{}) (result *turnpike.CallResult) {
	log.Printf("[rpc]%s: %v, %+v\n", URIXChatFetchChatList, args, kwargs)
	//_, user := getSessionFromDetails(kwargs["details"])

	//
	// chatType := args[0].(string)
	// chatTag := kwargs["tag"].(string)
	// TODO

	return nil
}

// 获取历史消息
func (m *Mid) fetchChatMsg(args []interface{}, kwargs map[string]interface{}) (result *turnpike.CallResult) {
	log.Printf("[rpc]%s: %v, %+v\n", URIXChatFetchChatMsgs, args, kwargs)
	_, user := getSessionFromDetails(kwargs["details"])
	chatID := uint64(args[0].(float64))
	channel, err := db.GetChannelByChatIDAndUser(chatID, user)
	if err != nil {
		return &turnpike.CallResult{Args: []interface{}{false, 1, err.Error()}}
	}

	var lid, rid uint64
	var limit int
	if kwargs["lid"] != nil {
		lid = uint64(kwargs["lid"].(float64))
	}
	if kwargs["rid"] != nil {
		rid = uint64(kwargs["rid"].(float64))
	}
	if kwargs["limit"] != nil {
		limit = int(kwargs["limit"].(float64))
	}

	msgs, err := ximHTTPClient.FetchChannelMsgs(channel, lid, rid, limit)
	if err != nil {
		return &turnpike.CallResult{Args: []interface{}{false, 1, err.Error()}}
	}

	chatMsg := ChatMsgs{
		ChatID: chatID,
		Msgs:   msgs,
	}
	return &turnpike.CallResult{Args: []interface{}{true, chatMsg}}
}

func (m *Mid) handleMsg(msg msgutils.Message) {
	switch x := msg.(type) {
	case *proto.Push:
		sessionID, err := m.xim.getSessionIDbyUID(x.UID)
		if err != nil {
			return
		}

		channel := x.Channel
		if chat, err := db.GetChatByChannel(channel); err == nil {
			chatMsg := ChatMsgs{
				ChatID: chat.ID,
				Msgs: []UserMsg{
					UserMsg{
						User: x.User,
						ID:   x.ID,
						Ts:   x.Ts,
						Msg:  x.Msg,
					},
				},
			}
			_ = m.xchat.Publish(fmt.Sprintf(URIXChatUserMsg, sessionID), []interface{}{chatMsg}, emptyKwargs)
		}
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
