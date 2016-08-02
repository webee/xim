package mid

import (
	"encoding/json"
	"strconv"
	"time"
	"xim/utils/nsutils"
	"xim/xchat/logic/db"
	pubtypes "xim/xchat/logic/pub/types"
	"xim/xchat/logic/service"
	"xim/xchat/logic/service/types"

	"gopkg.in/webee/turnpike.v2"
)

func getSessionFromID(sessionID interface{}) *Session {
	id := SessionID(sessionID.(turnpike.ID))
	if s, ok := GetSession(id); ok {
		return s
	}
	return nil
}

func getSessionFromDetails(d interface{}, forceCreate bool) *Session {
	details := d.(map[string]interface{})
	id := SessionID(details["session"].(turnpike.ID))
	if s, ok := GetSession(id); ok {
		return s
	}
	if forceCreate {
		ns, user := details["ns"].(string), details["user"].(string)
		return newSession(id, nsutils.EncodeNSUser(ns, user))
	}
	return nil
}

// 处理用户连接
func onJoin(args []interface{}, kwargs map[string]interface{}) {
	details := args[0].(map[string]interface{})
	s := getSessionFromDetails(details, true)
	AddSession(s)
	l.Debug("join: %s", s)
	// 上线状态
	arguments := &types.PubUserStatusArgs{
		InstanceID: instanceID,
		SessionID:  uint64(s.ID),
		User:       s.User,
		Status:     types.UserStatusOnline,
	}
	xchatLogic.AsyncCall(types.RPCXChatPubUserStatus, arguments)
}

// 处理用户断开注销
func onLeave(args []interface{}, kwargs map[string]interface{}) {
	id := SessionID(args[0].(turnpike.ID))
	s := RemoveSession(id)
	if s != nil {
		// 离开所有房间
		s.ExitAllRooms()
		l.Debug("left: %s", s)

		// 离线状态
		arguments := &types.PubUserStatusArgs{
			InstanceID: instanceID,
			SessionID:  uint64(s.ID),
			User:       s.User,
			Status:     types.UserStatusOffline,
			Info:       s.GetClientInfo(),
		}
		xchatLogic.AsyncCall(types.RPCXChatPubUserStatus, arguments)
	}
}

func doPubUserInfo(s *Session, infox interface{}) {
	info := ""
	switch x := infox.(type) {
	case string:
		info = x
	case map[string]interface{}:
		if s, err := json.Marshal(x); err != nil {
			info = string(s)
		} else {
			panic(err)
		}
	default:
		// panic.
		info = x.(string)
	}
	s.SetClientInfo(info)

	arguments := &types.PubUserStatusArgs{
		InstanceID: instanceID,
		SessionID:  uint64(s.ID),
		User:       s.User,
		Status:     types.UserStatusOnline,
		Info:       info,
	}
	xchatLogic.AsyncCall(types.RPCXChatPubUserStatus, arguments)
}

func pubUserInfo(s *Session, args []interface{}, kwargs map[string]interface{}) (rargs []interface{}, rkwargs map[string]interface{}, rerr APIError) {
	doPubUserInfo(s, args[0])
	return []interface{}{true}, nil, nil
}

func onPubUserInfo(s *Session, args []interface{}, kwargs map[string]interface{}) {
	doPubUserInfo(s, args[0])
}

// 用户发送消息, 会话消息
func sendMsg(s *Session, args []interface{}, kwargs map[string]interface{}) (rargs []interface{}, rkwargs map[string]interface{}, rerr APIError) {
	chatIdentity, err := service.ParseChatIdentity(args[0].(string))
	if err != nil {
		rerr = InvalidArgumentError
		return
	}
	chatID := chatIdentity.ID
	chatType := chatIdentity.Type

	msg := args[1].(string)
	if len(msg) > 64*1024 {
		rerr = MsgSizeExceedLimitError
		return
	}

	src := &pubtypes.MsgSource{
		InstanceID: instanceID,
		SessionID:  uint64(s.ID),
	}
	var message pubtypes.ChatMessage
	if err := xchatLogic.Call(types.RPCXChatSendMsg, &types.SendMsgArgs{
		Source:   src,
		ChatID:   chatID,
		ChatType: chatType,
		User:     s.User,
		Msg:      msg,
		Kind:     types.MsgKindChat,
	}, &message); err != nil {
		l.Warning("%s error: %s", types.RPCXChatSendMsg, err)
		rerr = newDefaultAPIError(err.Error())
		return
	}

	go push(src, &message)

	withMsg := false
	if x, ok := kwargs["with_msg"]; ok {
		withMsg = x.(bool)
	}
	if withMsg {
		toPushMsg := NewMessageFromPubMsg(&message)
		return []interface{}{true, message.ID, message.Ts}, map[string]interface{}{"msg": toPushMsg}, nil
	}

	return []interface{}{true, message.ID, message.Ts}, nil, nil
}

// 用户发布消息, 通知消息
func onPubMsg(s *Session, args []interface{}, kwargs map[string]interface{}) {
	chatIdentity, err := service.ParseChatIdentity(args[0].(string))
	if err != nil {
		return
	}
	chatID := chatIdentity.ID
	chatType := chatIdentity.Type
	msg := args[1].(string)
	if len(msg) > 16*1024 {
		// NOTE: msg exceed size limit
		return
	}

	src := &pubtypes.MsgSource{
		InstanceID: instanceID,
		SessionID:  uint64(s.ID),
	}
	xchatLogic.AsyncCall(types.RPCXChatSendMsg, &types.SendMsgArgs{
		Source:   src,
		ChatID:   chatID,
		ChatType: chatType,
		User:     s.User,
		Msg:      msg,
		Kind:     types.MsgKindChatNotify,
	})
}

// 创建会话
func newChat(s *Session, args []interface{}, kwargs map[string]interface{}) (rargs []interface{}, rkwargs map[string]interface{}, rerr APIError) {
	chatType := args[0].(string)
	if chatType != "user" && chatType != "group" && chatType != "self" {
		rerr = InvalidChatTypeError
		return
	}

	users := []string{s.User}
	for _, u := range args[1].([]interface{}) {
		users = append(users, u.(string))
	}
	title := args[2].(string)
	ext := ""
	if x, ok := kwargs["ext"]; ok {
		ext = x.(string)
	}

	xchatID, err := xchatHTTPClient.NewChat(chatType, users, title, "user", ext)
	if err != nil {
		rerr = newDefaultAPIError(err.Error())
		return
	}

	chatIdentity, err := service.ParseChatIdentity(xchatID)
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
		rerr = newDefaultAPIError(err.Error())
		return
	}
	return []interface{}{true, &userChat}, nil, nil
}

// 获取会话信息
func fetchChat(s *Session, args []interface{}, kwargs map[string]interface{}) (rargs []interface{}, rkwargs map[string]interface{}, rerr APIError) {
	chatIdentity, err := service.ParseChatIdentity(args[0].(string))
	if err != nil {
		rerr = InvalidArgumentError
		return
	}
	chatID := chatIdentity.ID
	chatType := chatIdentity.Type

	var userChat *db.UserChat
	if chatType == types.ChatTypeRoom {
		// fetch chat.
		chat := db.Chat{}
		if err := xchatLogic.Call(types.RPCXChatFetchChat, chatID, &chat); err != nil {
			rerr = newDefaultAPIError(err.Error())
			return
		}

		userChat = chatToUserChat(s.User, &chat)
	} else {
		// fetch user chat.
		uc := db.UserChat{}
		if err := xchatLogic.Call(types.RPCXChatFetchUserChat, &types.FetchUserChatArgs{
			User:   s.User,
			ChatID: chatID,
		}, &uc); err != nil {
			rerr = newDefaultAPIError(err.Error())
			return
		}
		userChat = &uc
	}
	return []interface{}{true, userChat}, nil, nil
}

// 获取会话成员信息
func fetchChatMembers(s *Session, args []interface{}, kwargs map[string]interface{}) (rargs []interface{}, rkwargs map[string]interface{}, rerr APIError) {
	chatIdentity, err := service.ParseChatIdentity(args[0].(string))
	if err != nil {
		rerr = InvalidArgumentError
		return
	}
	chatID := chatIdentity.ID
	//chatType := chatIdentity.Type

	// fetch members
	members := []db.Member{}
	if err := xchatLogic.Call(types.RPCXChatFetchUserChatMembers, &types.FetchUserChatMembersArgs{
		User:   s.User,
		ChatID: chatID,
	}, &members); err != nil {
		rerr = newDefaultAPIError(err.Error())
		return
	}
	return []interface{}{true, members}, nil, nil
}

// 获取会话列表
func fetchChatList(s *Session, args []interface{}, kwargs map[string]interface{}) (rargs []interface{}, rkwargs map[string]interface{}, rerr APIError) {
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
		rerr = newDefaultAPIError(err.Error())
		return
	}
	return []interface{}{true, userChats}, nil, nil
}

// 同时会话接收消息
func syncChatRecv(s *Session, args []interface{}, kwargs map[string]interface{}) (rargs []interface{}, rkwargs map[string]interface{}, rerr APIError) {
	chatIdentity, err := service.ParseChatIdentity(args[0].(string))
	if err != nil {
		rerr = InvalidArgumentError
	}
	if chatIdentity.Type == "room" {
		// 直接成功
		return []interface{}{true}, nil, nil
	}

	chatID := chatIdentity.ID
	msgID := uint64(args[1].(float64))

	// sync chat recv.
	if err := xchatLogic.Call(types.RPCXChatSyncUserChatRecv, &types.SyncUserChatRecvArgs{
		User:   s.User,
		ChatID: chatID,
		MsgID:  msgID,
	}, nil); err != nil {
		rerr = newDefaultAPIError(err.Error())
		return
	}
	return []interface{}{true}, nil, nil
}

// 获取历史消息
func fetchChatMsgs(s *Session, args []interface{}, kwargs map[string]interface{}) (rargs []interface{}, rkwargs map[string]interface{}, rerr APIError) {
	// params
	chatIdentity, err := service.ParseChatIdentity(args[0].(string))
	if err != nil {
		rerr = InvalidArgumentError
		return
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
		return []interface{}{true, []interface{}{}}, nil, nil
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
		rerr = newDefaultAPIError(err.Error())
		return
	}

	toPushMsgs := []*Message{}
	for _, msg := range msgs {
		toPushMsgs = append(toPushMsgs, NewMessageFromPubMsg(&msg))
	}
	return []interface{}{true, toPushMsgs}, nil, nil
}

// helplers
func chatToUserChat(user string, chat *db.Chat) *db.UserChat {
	return &db.UserChat{
		ID:      chat.ID,
		Type:    chat.Type,
		Title:   chat.Title,
		Tag:     chat.Tag,
		MsgID:   chat.MsgID,
		Created: chat.Created,
		Updated: chat.Created,
		User:    user,
		CurID:   chat.MsgID,
		Joined:  time.Now(),
	}
}

// 房间
// 进入房间
func enterRoom(s *Session, args []interface{}, kwargs map[string]interface{}) (rargs []interface{}, rkwargs map[string]interface{}, rerr APIError) {
	roomID, err := strconv.ParseUint(args[0].(string), 10, 64)
	if err != nil {
		rerr = InvalidArgumentError
		return
	}

	chatID, err := s.EnterRoom(roomID)
	if err != nil {
		rerr = newDefaultAPIError(err.Error())
		return
	}

	// fetch chat.
	chat := db.Chat{}
	if err := xchatLogic.Call(types.RPCXChatFetchChat, chatID, &chat); err != nil {
		rerr = newDefaultAPIError(err.Error())
		return
	}

	userChat := chatToUserChat(s.User, &chat)
	return []interface{}{true, userChat}, nil, nil
}

// 离开房间
func exitRoom(s *Session, args []interface{}, kwargs map[string]interface{}) (rargs []interface{}, rkwargs map[string]interface{}, rerr APIError) {
	roomID, err := strconv.ParseUint(args[0].(string), 10, 64)
	if err != nil {
		rerr = InvalidArgumentError
		return
	}

	chatIdentity, err := service.ParseChatIdentity(args[1].(string))
	if err != nil {
		rerr = newDefaultAPIError(err.Error())
		return
	}
	chatID := chatIdentity.ID

	s.ExitRoom(roomID, chatID)

	return []interface{}{true}, nil, nil
}

func joinChat(s *Session, args []interface{}, kwargs map[string]interface{}) (rargs []interface{}, rkwargs map[string]interface{}, rerr APIError) {
	chatIdentity, err := service.ParseChatIdentity(args[0].(string))
	if err != nil {
		rerr = InvalidArgumentError
		return
	}
	chatID := chatIdentity.ID
	chatType := chatIdentity.Type

	if err := xchatLogic.Call(types.RPCXChatJoinChat, &types.JoinChatArgs{
		ChatID:   chatID,
		ChatType: chatType,
		User:     s.User,
	}, nil); err != nil {
		rerr = newDefaultAPIError(err.Error())
		return
	}

	return []interface{}{true}, nil, nil
}

func exitChat(s *Session, args []interface{}, kwargs map[string]interface{}) (rargs []interface{}, rkwargs map[string]interface{}, rerr APIError) {
	chatIdentity, err := service.ParseChatIdentity(args[0].(string))
	if err != nil {
		rerr = InvalidArgumentError
		return
	}
	chatID := chatIdentity.ID
	chatType := chatIdentity.Type

	if err := xchatLogic.Call(types.RPCXChatExitChat, &types.ExitChatArgs{
		ChatID:   chatID,
		ChatType: chatType,
		User:     s.User,
	}, nil); err != nil {
		rerr = newDefaultAPIError(err.Error())
		return
	}

	return []interface{}{true}, nil, nil
}

// 客服
// 获取我的客服会话
func getCsChat(s *Session, args []interface{}, kwargs map[string]interface{}) (rargs []interface{}, rkwargs map[string]interface{}, rerr APIError) {
	xchatID, err := xchatHTTPClient.NewChat("cs", []string{s.User}, "我的客服", "_cs", "")
	if err != nil {
		rerr = newDefaultAPIError(err.Error())
		return
	}

	chatIdentity, err := service.ParseChatIdentity(xchatID)
	if err != nil {
		rerr = InvalidArgumentError
		return
	}
	chatID := chatIdentity.ID

	// fetch user chat.
	userChat := db.UserChat{}
	if err := xchatLogic.Call(types.RPCXChatFetchUserChat, &types.FetchUserChatArgs{
		User:   s.User,
		ChatID: chatID,
	}, &userChat); err != nil {
		rerr = newDefaultAPIError(err.Error())
		return
	}
	return []interface{}{true, &userChat}, nil, nil
}
