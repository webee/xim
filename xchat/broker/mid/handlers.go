package mid

import (
	"xim/xchat/logic/db"
	pubtypes "xim/xchat/logic/pub/types"
	"xim/xchat/logic/service/types"

	"gopkg.in/jcelliott/turnpike.v2"
)

func getSessionFromDetails(d interface{}, forceCreate bool) *Session {
	details := d.(map[string]interface{})
	id := SessionID(details["session"].(turnpike.ID))
	if s, ok := GetSession(id); ok {
		return s
	}
	if forceCreate {
		return newSession(id, details["user"].(string))
	}
	return nil
}

// 处理用户连接
func onJoin(args []interface{}, kwargs map[string]interface{}) {
	details := args[0].(map[string]interface{})
	s := getSessionFromDetails(details, true)
	AddSession(s)
	l.Debug("join: %s", s)
}

// 处理用户断开注销
func onLeave(args []interface{}, kwargs map[string]interface{}) {
	id := SessionID(args[0].(turnpike.ID))
	s := RemoveSession(id)
	l.Debug("left: %s", s)
}

// ping
func ping(args []interface{}, kwargs map[string]interface{}) (result *turnpike.CallResult) {
	l.Debug("[rpc]%s: %v, %+v\n", URIXChatPing, args, kwargs)
	s := getSessionFromDetails(kwargs["details"], false)
	if s == nil {
		return &turnpike.CallResult{Args: []interface{}{false, 2, "session exception"}}
	}
	payloadSize := 0
	if len(args) > 0 {
		payloadSize = int(args[0].(float64))
	}

	if payloadSize < 0 {
		payloadSize = 0
	} else if payloadSize > 1024*1024 {
		payloadSize = 1024 * 1024
	}

	payload := []byte{}
	for i := 1; i < payloadSize; i++ {
		payload = append(payload, 0x31)
	}

	return &turnpike.CallResult{Args: []interface{}{true, s.ID, string(payload)}}
}

// 用户发送消息
func sendMsg(args []interface{}, kwargs map[string]interface{}) (result *turnpike.CallResult) {
	l.Debug("[rpc]%s: %v, %+v", URIXChatSendMsg, args, kwargs)
	s := getSessionFromDetails(kwargs["details"], false)
	if s == nil {
		return &turnpike.CallResult{Args: []interface{}{false, 2, "session exception"}}
	}

	chatID := uint64(args[0].(float64))
	msg := args[1].(string)

	var message pubtypes.Message
	if err := xchatLogic.Call(types.RPCXChatSendMsg, &types.SendMsgArgs{
		ChatID: chatID,
		User:   s.User,
		Msg:    msg,
	}, &message); err != nil {
		l.Warning("error: %s", err)
		return &turnpike.CallResult{Args: []interface{}{false, 1, err.Error()}}
	}
	// update sending id.
	pushSessMsg(s, &message)

	return &turnpike.CallResult{Args: []interface{}{true, message.ID, message.Ts}}
}

// 创建会话
func newChat(args []interface{}, kwargs map[string]interface{}) (result *turnpike.CallResult) {
	l.Debug("[rpc]%s: %v, %+v\n", URIXChatNewChat, args, kwargs)
	s := getSessionFromDetails(kwargs["details"], false)
	if s == nil {
		return &turnpike.CallResult{Args: []interface{}{false, 2, "session exception"}}
	}

	chatType := args[0].(string)
	users := []string{s.User}
	for _, u := range args[1].([]interface{}) {
		users = append(users, u.(string))
	}
	title := args[2].(string)

	chatID, err := xchatHTTPClient.NewChat(chatType, users, title)
	if err != nil {
		return &turnpike.CallResult{Args: []interface{}{false, 1, err.Error()}}
	}

	return &turnpike.CallResult{Args: []interface{}{true, chatID}}
}

// 获取会话信息
func fetchChat(args []interface{}, kwargs map[string]interface{}) (result *turnpike.CallResult) {
	l.Debug("[rpc]%s: %v, %+v\n", URIXChatFetchChat, args, kwargs)
	s := getSessionFromDetails(kwargs["details"], false)
	if s == nil {
		return &turnpike.CallResult{Args: []interface{}{false, 2, "session exception"}}
	}
	chatID := uint64(args[0].(float64))

	// fetch chat.
	chat := db.Chat{}
	if err := xchatLogic.Call(types.RPCXChatFetchChat, chatID, &chat); err != nil {
		return &turnpike.CallResult{Args: []interface{}{false, 1, err.Error()}}
	}
	return &turnpike.CallResult{Args: []interface{}{true, NewChatFromDBChat(&chat)}}
}

// 获取会话列表
func fetchChatList(args []interface{}, kwargs map[string]interface{}) (result *turnpike.CallResult) {
	l.Debug("[rpc]%s: %v, %+v\n", URIXChatChatList, args, kwargs)
	s := getSessionFromDetails(kwargs["details"], false)
	if s == nil {
		return &turnpike.CallResult{Args: []interface{}{false, 2, "session exception"}}
	}

	// fetch chat.
	chats := []db.UserChat{}
	if err := xchatLogic.Call(types.RPCXChatFetchUserChatList, s.User, &chats); err != nil {
		return &turnpike.CallResult{Args: []interface{}{false, 1, err.Error()}}
	}
	return &turnpike.CallResult{Args: []interface{}{true, chats}}
}

// 获取历史消息
func fetchChatMsg(args []interface{}, kwargs map[string]interface{}) (result *turnpike.CallResult) {
	l.Debug("[rpc]%s: %v, %+v\n", URIXChatFetchChatMsgs, args, kwargs)
	return nil
	// _, user := getSessionFromDetails(kwargs["details"])
	// chatID := uint64(args[0].(float64))
}

func sub(handler turnpike.EventHandler) turnpike.EventHandler {
	return func(args []interface{}, kargs map[string]interface{}) {
		defer func() {
			if r := recover(); r != nil {
				l.Warning("sub error: %s", r)
			}
		}()
		handler(args, kargs)
	}
}

func call(handler turnpike.BasicMethodHandler) turnpike.BasicMethodHandler {
	return func(args []interface{}, kargs map[string]interface{}) (result *turnpike.CallResult) {
		defer func() {
			if r := recover(); r != nil {
				l.Warning("call error: %s", r)
				result = &turnpike.CallResult{Err: turnpike.ErrInvalidArgument, Args: []interface{}{r}}
			}
		}()
		return handler(args, kargs)
	}
}
