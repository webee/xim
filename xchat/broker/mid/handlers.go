package mid

import (
	"xim/xchat/logic/db"
	"xim/xchat/logic/service"

	"gopkg.in/jcelliott/turnpike.v2"
)

func getSessionFromDetails(d interface{}) *Session {
	details := d.(map[string]interface{})
	return &Session{
		ID:   SessionID(details["session"].(turnpike.ID)),
		User: details["user"].(string),
	}
}

// 处理用户连接
func onJoin(args []interface{}, kwargs map[string]interface{}) {
	details := args[0].(map[string]interface{})
	s := getSessionFromDetails(details)
	AddSession(s)
	l.Debug("join: %s", s)
}

// 处理用户断开注销
func onLeave(args []interface{}, kwargs map[string]interface{}) {
	id := SessionID(args[0].(turnpike.ID))
	s := RemoveSession(id)
	l.Debug("left: %s", s)
}

// 用户发送消息
func sendMsg(args []interface{}, kwargs map[string]interface{}) (result *turnpike.CallResult) {
	l.Debug("[rpc]%s: %v, %+v\n", URIXChatSendMsg, args, kwargs)
	s := getSessionFromDetails(kwargs["details"])

	chatID := uint64(args[0].(float64))
	msg := args[1].(string)
	res, err := xchatDC.Call(service.XChat.MethodSendMsg, &service.SendMsgRequest{
		ChatID: chatID,
		User:   s.User,
		Msg:    msg,
	})
	if err != nil {
		return &turnpike.CallResult{Args: []interface{}{false, 1, err.Error()}}
	}
	message := (res).(*db.Message)

	return &turnpike.CallResult{Args: []interface{}{true, message.MsgID, message.Ts.Unix()}}
}

// 获取会话信息
func newChat(args []interface{}, kwargs map[string]interface{}) (result *turnpike.CallResult) {
	l.Debug("[rpc]%s: %v, %+v\n", URIXChatNewChat, args, kwargs)
	return nil
	// _, user := getSessionFromDetails(kwargs["details"])
}

// 获取会话列表
func fetchChatList(args []interface{}, kwargs map[string]interface{}) (result *turnpike.CallResult) {
	l.Debug("[rpc]%s: %v, %+v\n", URIXChatChatList, args, kwargs)
	return nil
	// _, user := getSessionFromDetails(kwargs["details"])
}

// 获取历史消息
func fetchChatMsg(args []interface{}, kwargs map[string]interface{}) (result *turnpike.CallResult) {
	l.Debug("[rpc]%s: %v, %+v\n", URIXChatFetchChatMsgs, args, kwargs)
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
