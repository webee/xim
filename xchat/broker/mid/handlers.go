package mid

import (
	"strconv"
	"time"
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
	if s != nil {
		// 离开所有房间
		s.ExitAllRooms()
		l.Debug("left: %s", s)

		clientInfo := s.GetClientInfo()
		if clientInfo != "" {
			arguments := &types.PubUserStatusArgs{
				User:   s.User,
				Status: types.UserStatusOffline,
				Info:   clientInfo,
			}
			xchatLogic.AsyncCall(types.RPCXChatPubUserStatus, arguments)
		}
	}
}

// ping
func ping(args []interface{}, kwargs map[string]interface{}) (result *turnpike.CallResult) {
	l.Debug("[rpc]%s: %v, %+v\n", URIXChatPing, args, kwargs)
	s := getSessionFromDetails(kwargs["details"], false)
	if s == nil {
		return &turnpike.CallResult{Args: []interface{}{false, 2, "session exception"}}
	}
	// TODO: 添加多种探测功能，rpc, 获取状态等
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

func onPubUserInfo(args []interface{}, kwargs map[string]interface{}) {
	l.Debug("[pub]%s: %v, %+v\n", URIXChatPubUserInfo, args, kwargs)
	s := getSessionFromDetails(kwargs["details"], false)
	if s == nil {
		return
	}

	info := args[0].(string)
	s.SetClientInfo(info)

	arguments := &types.PubUserStatusArgs{
		User:   s.User,
		Status: types.UserStatusOnline,
		Info:   info,
	}
	xchatLogic.AsyncCall(types.RPCXChatPubUserStatus, arguments)
}

// 用户发送消息, 会话消息
func sendMsg(args []interface{}, kwargs map[string]interface{}) (result *turnpike.CallResult) {
	l.Debug("[rpc]%s: %v, %+v", URIXChatSendMsg, args, kwargs)
	s := getSessionFromDetails(kwargs["details"], false)
	if s == nil {
		return &turnpike.CallResult{Args: []interface{}{false, 2, "session exception"}}
	}

	chatIdentity, err := ParseChatIdentity(args[0].(string))
	if err != nil {
		return &turnpike.CallResult{Args: []interface{}{false, 1, err.Error()}}
	}
	chatID := chatIdentity.ID
	chatType := chatIdentity.Type

	msg := args[1].(string)

	p := s.GetChatPustState(chatID)
	p.setSending()
	defer p.doneSending()

	var message pubtypes.ChatMessage
	if err := xchatLogic.Call(types.RPCXChatSendMsg, &types.SendMsgArgs{
		ChatID:   chatID,
		ChatType: chatType,
		User:     s.User,
		Msg:      msg,
		Kind:     types.MsgKindChat,
	}, &message); err != nil {
		l.Warning("%s error: %s", types.RPCXChatSendMsg, err)
		return &turnpike.CallResult{Args: []interface{}{false, 1, err.Error()}}
	}
	// update sending id.
	pushSessMsg(p, &message)

	return &turnpike.CallResult{Args: []interface{}{true, message.ID, message.Ts}}
}

// 用户发布消息, 通知消息
func onPubMsg(args []interface{}, kwargs map[string]interface{}) {
	l.Debug("[pub]%s: %v, %+v", URIXChatPubMsg, args, kwargs)
	s := getSessionFromDetails(kwargs["details"], false)
	if s == nil {
		return
	}

	chatIdentity, err := ParseChatIdentity(args[0].(string))
	if err != nil {
		return
	}
	chatID := chatIdentity.ID
	chatType := chatIdentity.Type
	msg := args[1].(string)

	xchatLogic.AsyncCall(types.RPCXChatSendMsg, &types.SendMsgArgs{
		ChatID:   chatID,
		ChatType: chatType,
		User:     s.User,
		Msg:      msg,
		Kind:     types.MsgKindChatNotify,
	})
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

	xchatID, err := xchatHTTPClient.NewChat(chatType, users, title, "user")
	if err != nil {
		return &turnpike.CallResult{Args: []interface{}{false, 1, err.Error()}}
	}

	chatIdentity, err := ParseChatIdentity(xchatID)
	if err != nil {
		return
	}
	chatID := chatIdentity.ID

	// fetch user chat.
	userChat := db.UserChat{}
	if err := xchatLogic.Call(types.RPCXChatFetchUserChat, &types.FetchUserChatArgs{
		User:   s.User,
		ChatID: chatID,
	}, &userChat); err != nil {
		return &turnpike.CallResult{Args: []interface{}{false, 1, err.Error()}}
	}
	return &turnpike.CallResult{Args: []interface{}{true, &userChat}}
}

// 获取会话信息
func fetchChat(args []interface{}, kwargs map[string]interface{}) (result *turnpike.CallResult) {
	l.Debug("[rpc]%s: %v, %+v\n", URIXChatFetchChat, args, kwargs)
	s := getSessionFromDetails(kwargs["details"], false)
	if s == nil {
		return &turnpike.CallResult{Args: []interface{}{false, 2, "session exception"}}
	}
	chatIdentity, err := ParseChatIdentity(args[0].(string))
	if err != nil {
		return &turnpike.CallResult{Args: []interface{}{false, 1, err.Error()}}
	}
	chatID := chatIdentity.ID

	// fetch user chat.
	userChat := db.UserChat{}
	if err := xchatLogic.Call(types.RPCXChatFetchUserChat, &types.FetchUserChatArgs{
		User:   s.User,
		ChatID: chatID,
	}, &userChat); err != nil {
		return &turnpike.CallResult{Args: []interface{}{false, 1, err.Error()}}
	}
	return &turnpike.CallResult{Args: []interface{}{true, &userChat}}
}

// 获取会话列表
func fetchChatList(args []interface{}, kwargs map[string]interface{}) (result *turnpike.CallResult) {
	l.Debug("[rpc]%s: %v, %+v\n", URIXChatFetchChatList, args, kwargs)
	s := getSessionFromDetails(kwargs["details"], false)
	if s == nil {
		return &turnpike.CallResult{Args: []interface{}{false, 2, "session exception"}}
	}

	// params
	var onlyUnsync bool
	if x, ok := kwargs["only_unsync"]; ok {
		onlyUnsync = x.(bool)
	}

	// fetch chat.
	userChats := []db.UserChat{}
	if err := xchatLogic.Call(types.RPCXChatFetchUserChatList, &types.FetchUserChatListArgs{
		User:       s.User,
		OnlyUnsync: onlyUnsync,
	}, &userChats); err != nil {
		return &turnpike.CallResult{Args: []interface{}{false, 1, err.Error()}}
	}
	return &turnpike.CallResult{Args: []interface{}{true, userChats}}
}

// 同时会话接收消息
func syncChatRecv(args []interface{}, kwargs map[string]interface{}) (result *turnpike.CallResult) {
	l.Debug("[rpc]%s: %v, %+v\n", URIXChatSyncChatRecv, args, kwargs)
	s := getSessionFromDetails(kwargs["details"], false)
	if s == nil {
		return &turnpike.CallResult{Args: []interface{}{false, 2, "session exception"}}
	}
	chatIdentity, err := ParseChatIdentity(args[0].(string))
	if err != nil {
		return &turnpike.CallResult{Args: []interface{}{false, 1, err.Error()}}
	}
	chatID := chatIdentity.ID
	msgID := uint64(args[1].(float64))

	// sync chat recv.
	if err := xchatLogic.Call(types.RPCXChatSyncUserChatRecv, &types.SyncUserChatRecvArgs{
		User:   s.User,
		ChatID: chatID,
		MsgID:  msgID,
	}, nil); err != nil {
		return &turnpike.CallResult{Args: []interface{}{false, 1, err.Error()}}
	}
	return &turnpike.CallResult{Args: []interface{}{true}}
}

// 获取历史消息
func fetchChatMsgs(args []interface{}, kwargs map[string]interface{}) (result *turnpike.CallResult) {
	l.Debug("[rpc]%s: %v, %+v\n", URIXChatFetchChatMsgs, args, kwargs)
	s := getSessionFromDetails(kwargs["details"], false)
	if s == nil {
		return &turnpike.CallResult{Args: []interface{}{false, 2, "session exception"}}
	}
	// params
	chatIdentity, err := ParseChatIdentity(args[0].(string))
	if err != nil {
		return &turnpike.CallResult{Args: []interface{}{false, 1, err.Error()}}
	}
	chatID := chatIdentity.ID
	chatType := chatIdentity.Type

	var lid, rid uint64
	var limit int
	var desc bool
	if kwargs["lid"] != nil {
		lid = uint64(kwargs["lid"].(float64))
	} else {
		desc = true
	}

	if kwargs["rid"] != nil {
		rid = uint64(kwargs["rid"].(float64))
	}

	if lid > 0 && lid+1 >= rid {
		return &turnpike.CallResult{Args: []interface{}{true, []interface{}{}}}
	}

	if kwargs["limit"] != nil {
		limit = int(kwargs["limit"].(float64))
	}
	if limit <= 0 {
		limit = 150
	} else if limit > 1000 {
		limit = 1000
	}

	var msgs []pubtypes.ChatMessage
	arguments := &types.FetchUserChatMessagesArgs{
		User:     s.User,
		ChatID:   chatID,
		ChatType: chatType,
		LID:      lid,
		RID:      rid,
		Limit:    limit,
		Desc:     desc,
	}

	if err := xchatLogic.Call(types.RPCXChatFetchUserChatMessages, arguments, &msgs); err != nil {
		return &turnpike.CallResult{Args: []interface{}{false, 1, err.Error()}}
	}

	toPushMsgs := []*Message{}
	for _, msg := range msgs {
		toPushMsgs = append(toPushMsgs, NewMessageFromDBMsg(&msg))
	}
	return &turnpike.CallResult{Args: []interface{}{true, toPushMsgs}}
}

// 房间
// 进入房间
func enterRoom(args []interface{}, kwargs map[string]interface{}) (result *turnpike.CallResult) {
	l.Debug("[rpc]%s: %v, %+v\n", URIXChatEnterRoom, args, kwargs)
	s := getSessionFromDetails(kwargs["details"], false)
	if s == nil {
		return &turnpike.CallResult{Args: []interface{}{false, 2, "session exception"}}
	}

	roomID, err := strconv.ParseUint(args[0].(string), 10, 64)
	if err != nil {
		return &turnpike.CallResult{Args: []interface{}{false, 1, err.Error()}}
	}

	chatID, err := s.EnterRoom(roomID)
	if err != nil {
		return &turnpike.CallResult{Args: []interface{}{false, 1, err.Error()}}
	}

	// fetch chat.
	chat := db.Chat{}
	if err := xchatLogic.Call(types.RPCXChatFetchChat, chatID, &chat); err != nil {
		return &turnpike.CallResult{Args: []interface{}{false, 1, err.Error()}}
	}

	userChat := db.UserChat{
		ID:      chat.ID,
		Type:    chat.Type,
		Title:   chat.Title,
		Tag:     chat.Tag,
		MsgID:   chat.MsgID,
		Created: chat.Created,
		Updated: chat.Created,
		User:    s.User,
		CurID:   chat.MsgID,
		Joined:  time.Now(),
	}

	return &turnpike.CallResult{Args: []interface{}{true, &userChat}}
}

// 离开房间
func exitRoom(args []interface{}, kwargs map[string]interface{}) (result *turnpike.CallResult) {
	l.Debug("[rpc]%s: %v, %+v\n", URIXChatEnterRoom, args, kwargs)
	s := getSessionFromDetails(kwargs["details"], false)
	if s == nil {
		return &turnpike.CallResult{Args: []interface{}{false, 2, "session exception"}}
	}

	roomID, err := strconv.ParseUint(args[0].(string), 10, 64)
	if err != nil {
		return &turnpike.CallResult{Args: []interface{}{false, 1, err.Error()}}
	}

	chatIdentity, err := ParseChatIdentity(args[1].(string))
	if err != nil {
		return &turnpike.CallResult{Args: []interface{}{false, 1, err.Error()}}
	}
	chatID := chatIdentity.ID

	s.ExitRoom(roomID, chatID)

	return &turnpike.CallResult{Args: []interface{}{true}}
}

// 客服
// 获取我的客服会话
func getCsChat(args []interface{}, kwargs map[string]interface{}) (result *turnpike.CallResult) {
	l.Debug("[rpc]%s: %v, %+v\n", URIXChatGetCsChat, args, kwargs)
	s := getSessionFromDetails(kwargs["details"], false)
	if s == nil {
		return &turnpike.CallResult{Args: []interface{}{false, 2, "session exception"}}
	}

	xchatID, err := xchatHTTPClient.NewChat("cs", []string{s.User}, "我的客服", "_cs")
	if err != nil {
		return &turnpike.CallResult{Args: []interface{}{false, 1, err.Error()}}
	}

	chatIdentity, err := ParseChatIdentity(xchatID)
	if err != nil {
		return &turnpike.CallResult{Args: []interface{}{false, 1, err.Error()}}
	}
	chatID := chatIdentity.ID

	// fetch user chat.
	userChat := db.UserChat{}
	if err := xchatLogic.Call(types.RPCXChatFetchUserChat, &types.FetchUserChatArgs{
		User:   s.User,
		ChatID: chatID,
	}, &userChat); err != nil {
		return &turnpike.CallResult{Args: []interface{}{false, 1, err.Error()}}
	}
	return &turnpike.CallResult{Args: []interface{}{true, &userChat}}
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
				result = &turnpike.CallResult{Err: turnpike.ErrInvalidArgument, Args: []interface{}{false, 4, turnpike.ErrInvalidArgument}}
			}
		}()
		return handler(args, kargs)
	}
}
