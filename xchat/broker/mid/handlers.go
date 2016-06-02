package mid

import (
	"xim/xchat/logic/service"

	"gopkg.in/jcelliott/turnpike.v2"
)

// 处理用户连接
func onJoin(args []interface{}, kwargs map[string]interface{}) {
	details := args[0].(map[string]interface{})
	l.Debug("join: %+v", details)
}

// 处理用户断开注销
func onLeave(args []interface{}, kwargs map[string]interface{}) {
	sessionID := uint64(args[0].(turnpike.ID))
	l.Debug("<%d> left\n", sessionID)
}

func getSessionFromDetails(d interface{}) (sessionID uint64, user string) {
	details := d.(map[string]interface{})
	sessionID = uint64(details["session"].(turnpike.ID))
	user = details["user"].(string)
	return
}

// 用户发送消息
func sendMsg(args []interface{}, kwargs map[string]interface{}) (result *turnpike.CallResult) {
	l.Debug("[rpc]%s: %v, %+v\n", URIXChatSendMsg, args, kwargs)
	_, user := getSessionFromDetails(kwargs["details"])

	chatID := uint64(args[0].(float64))
	msg := args[1].(string)
	res, err := xchatDC.Call(service.XChat.MethodSendMsg, &service.SendMsgRequest{
		ChatID: chatID,
		User:   user,
		Msg:    msg,
	})
	if err != nil {
		return &turnpike.CallResult{Args: []interface{}{false, 1, err.Error()}}
	}
	reply := (res).(*service.SendMsgReply)

	return &turnpike.CallResult{Args: []interface{}{true, reply.MsgID, reply.Ts.Unix()}}
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
