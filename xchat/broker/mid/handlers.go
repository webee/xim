package mid

import (
	"fmt"
	"xim/xchat/logic/db"
	"xim/xchat/logic/rpcservice/types"

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

	var message *db.Message
	if err := xchatLogic.Call(types.RPCXChatSendMsg, &types.SendMsgArgs{
		ChatID: chatID,
		User:   s.User,
		Msg:    msg,
	}, &message); err != nil {
		l.Warning("error: %s", err)
		return &turnpike.CallResult{Args: []interface{}{false, 1, err.Error()}}
	}

	// push
	go func() {
		var members []db.Member
		if err := xchatLogic.Call(types.RPCXChatFetchChatMembers, chatID, &members); err != nil {
			l.Warning("fetch chat[%d] members error: %s", chatID, err)
			return
		}
		toPushMsg := NewMessageFromDBMsg(message)
		for _, member := range members {
			ss := GetUserSessions(member.User)
			for _, x := range ss {
				if x.ID == s.ID {
					continue
				}
				_ = xchat.Publish(fmt.Sprintf(URIXChatUserMsg, x.ID), []interface{}{toPushMsg}, emptyKwargs)
			}
		}
	}()

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

func sub(handler turnpike.EventHandler) turnpike.EventHandler {
	return func(args []interface{}, kargs map[string]interface{}) {
		defer func() {
			if r := recover(); r != nil {
				l.Emergency("sub error: %s", r)
			}
		}()
		handler(args, kargs)
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
