package mid

import (
	"log"
	"xim/xchat/broker/router"

	"gopkg.in/jcelliott/turnpike.v2"
)

// Mid is the mid of router and xim.
type Mid struct {
	xchat  *turnpike.Client
	config *Config
}

var (
	mid         *Mid
	emptyArgs   = []interface{}{}
	emptyKwargs = make(map[string]interface{})
)

// Setup initialze mid.
func Setup(config *Config, xchatRouter *router.XChatRouter) {
	xchat, err := xchatRouter.GetLocalClient("xchat", nil)
	if err != nil {
		log.Fatalln("create xchat error:", err)
	}

	mid = &Mid{
		xchat:  xchat,
		config: config,
	}

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

	if err := xchat.BasicRegister(URIXChatSendMsg, call(m.sendMsg)); err != nil {
		log.Fatalf("Error register %s: %s\n", URIXChatSendMsg, err)
	}

	if err := xchat.BasicRegister(URIXChatFetchChatMsgs, call(m.fetchChatMsg)); err != nil {
		log.Fatalf("Error register %s: %s\n", URIXChatFetchChatMsgs, err)
	}

	if err := xchat.BasicRegister(URIXChatNewChat, call(m.newChat)); err != nil {
		log.Fatalf("Error register %s: %s\n", URIXChatNewChat, err)
	}

	if err := xchat.BasicRegister(URIXChatFetchChatList, call(m.fetchChatList)); err != nil {
		log.Fatalf("Error register %s: %s\n", URIXChatNewChat, err)
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
	log.Printf("<%d> left\n", sessionID)
}

func getSessionFromDetails(d interface{}) (sessionID uint64, user string) {
	details := d.(map[string]interface{})
	sessionID = uint64(details["session"].(turnpike.ID))
	user = details["user"].(string)
	return
}

// 用户发送消息
func (m *Mid) sendMsg(args []interface{}, kwargs map[string]interface{}) (result *turnpike.CallResult) {
	log.Printf("[rpc]%s: %v, %+v\n", URIXChatSendMsg, args, kwargs)
	return nil
	// sessionID, user := getSessionFromDetails(kwargs["details"])
	//
	// chatID := uint64(args[0].(float64))
	// msg := args[1]
	// return &turnpike.CallResult{Args: []interface{}{true, id, ts}}
}

// 获取会话信息
func (m *Mid) newChat(args []interface{}, kwargs map[string]interface{}) (result *turnpike.CallResult) {
	log.Printf("[rpc]%s: %v, %+v\n", URIXChatNewChat, args, kwargs)
	return nil
	// _, user := getSessionFromDetails(kwargs["details"])
}

// 获取会话列表
func (m *Mid) fetchChatList(args []interface{}, kwargs map[string]interface{}) (result *turnpike.CallResult) {
	log.Printf("[rpc]%s: %v, %+v\n", URIXChatChatList, args, kwargs)
	return nil
	// _, user := getSessionFromDetails(kwargs["details"])
}

// 获取历史消息
func (m *Mid) fetchChatMsg(args []interface{}, kwargs map[string]interface{}) (result *turnpike.CallResult) {
	log.Printf("[rpc]%s: %v, %+v\n", URIXChatFetchChatMsgs, args, kwargs)
	return nil
	// _, user := getSessionFromDetails(kwargs["details"])
	// chatID := uint64(args[0].(float64))
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
