package mid

import (
	"encoding/json"
	"strconv"
	"time"
	"xim/utils/nsutils"
	"xim/xchat/logic/db"
	pubtypes "xim/xchat/logic/pub/types"
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

func getSessionFromSessionDetails(sessionID turnpike.ID, details map[string]interface{}, forceCreate bool) *Session {
	id := SessionID(sessionID)
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
	sessionID := args[0].(turnpike.ID)
	details := args[1].(map[string]interface{})
	if _, ok := details["is_local"]; ok {
		// ignore local client.
		return
	}

	s := getSessionFromSessionDetails(sessionID, details, true)
	if s == nil {
		return
	}

	AddSession(s)
	l.Debug("join: %s", s)
	// 上线状态
	doPubUserStatus(s, types.UserStatusOnline)
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
		doPubUserStatus(s, types.UserStatusOffline)
		doPubUserInfo(s, types.UserStatusOffline, s.GetClientInfo())
	}
}

func doPubUserStatus(s *Session, infox interface{}) {
	switch x := infox.(type) {
	case string:
		// set status
		arguments := &types.PubUserStatusArgs{
			InstanceID: instanceID,
			SessionID:  uint64(s.ID),
			User:       s.User,
			Status:     x,
		}
		xchatLogic.AsyncCall(types.RPCXChatPubUserStatus, arguments)
	}
}

func pubUserStatusInfo(s *Session, args []interface{}, kwargs map[string]interface{}) (rargs []interface{}, rkwargs map[string]interface{}, rerr APIError) {
	doPubUserStatus(s, args[0])
	return []interface{}{true}, nil, nil
}

func onPubUserStatusInfo(s *Session, args []interface{}, kwargs map[string]interface{}) {
	doPubUserStatus(s, args[0])
}

func doPubUserInfo(s *Session, status string, infox interface{}) {
	info := ""
	switch x := infox.(type) {
	case string:
		info = x
	case map[string]interface{}:
		if s, err := json.Marshal(x); err == nil {
			info = string(s)
		} else {
			panic(err)
		}
	default:
		// panic.
		info = x.(string)
	}
	s.SetClientInfo(info)

	arguments := &types.PubUserInfoArgs{
		PubUserStatusArgs: types.PubUserStatusArgs{
			InstanceID: instanceID,
			SessionID:  uint64(s.ID),
			User:       s.User,
			Status:     status,
		},
		Info: info,
	}
	xchatLogic.AsyncCall(types.RPCXChatPubUserInfo, arguments)
}

func pubUserInfo(s *Session, args []interface{}, kwargs map[string]interface{}) (rargs []interface{}, rkwargs map[string]interface{}, rerr APIError) {
	doPubUserInfo(s, types.UserStatusOnline, args[0])
	return []interface{}{true}, nil, nil
}

func onPubUserInfo(s *Session, args []interface{}, kwargs map[string]interface{}) {
	doPubUserInfo(s, types.UserStatusOnline, args[0])
}

func bindSendMsgArgs(s *Session, args []interface{}) (sendMsgArgs *types.SendMsgArgs, rerr APIError) {
	chatIdentity, err := db.ParseChatIdentity(args[0].(string))
	if err != nil {
		return nil, InvalidArgumentError
	}
	chatID := chatIdentity.ID
	chatType := chatIdentity.Type

	msg := args[1].(string)
	if len(msg) > 64*1024 {
		// NOTE: msg exceed size limit
		return nil, MsgSizeExceedLimitError
	}
	domain := ""
	if len(args) >= 3 {
		domain = args[2].(string)
	}

	src := &pubtypes.MsgSource{
		InstanceID: instanceID,
		SessionID:  uint64(s.ID),
	}

	sendMsgArgs = &types.SendMsgArgs{
		Source:   src,
		ChatID:   chatID,
		ChatType: chatType,
		Domain:   domain,
		User:     s.User,
		Msg:      msg,
	}
	return
}

func bindSendUserMsgArgs(s *Session, args []interface{}, kwargs map[string]interface{}) (sendMsgArgs *types.SendUserMsgArgs, rerr APIError) {
	toUser := args[0].(string)
	isNsUser := false
	if x, ok := kwargs["is_ns_user"]; ok {
		isNsUser = x.(bool)
	}

	ns, _ := nsutils.DecodeNSUser(s.User)
	if !isNsUser {
		toUser = nsutils.EncodeNSUser(ns, toUser)
	}

	msg := args[1].(string)
	if len(msg) > 64*1024 {
		// NOTE: msg exceed size limit
		return nil, MsgSizeExceedLimitError
	}
	domain := ""
	if len(args) >= 3 {
		domain = args[2].(string)
	}

	src := &pubtypes.MsgSource{
		InstanceID: instanceID,
		SessionID:  uint64(s.ID),
	}

	var options *types.SendMsgOptions
	if ns == "test" && isNsUser {
		toNs, _ := nsutils.DecodeNSUser(toUser)
		if toNs == "test" {
			// test user to test user, ignore perm check.
			options = &types.SendMsgOptions{
				IgnorePermCheck: true,
			}
		}
	}

	sendMsgArgs = &types.SendUserMsgArgs{
		Source:  src,
		ToUser:  toUser,
		Domain:  domain,
		User:    s.User,
		Msg:     msg,
		Options: options,
	}
	return
}

// 用户发送消息, 会话消息
func sendMsg(s *Session, args []interface{}, kwargs map[string]interface{}) (rargs []interface{}, rkwargs map[string]interface{}, rerr APIError) {
	sendMsgArgs, rerr := bindSendMsgArgs(s, args)
	if rerr != nil {
		return
	}

	var message pubtypes.ChatMessage
	if err := xchatLogic.Call(types.RPCXChatSendMsg, sendMsgArgs, &message); err != nil {
		l.Warning("%s error: %s", types.RPCXChatSendMsg, err)
		rerr = newDefaultAPIError(err.Error())
		return
	}

	go push(sendMsgArgs.Source, &message)

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
func onPubNotify(s *Session, args []interface{}, kwargs map[string]interface{}) {
	sendMsgArgs, rerr := bindSendMsgArgs(s, args)
	if rerr != nil {
		return
	}

	xchatLogic.AsyncCall(types.RPCXChatSendNotify, sendMsgArgs)
}

func sendNotify(s *Session, args []interface{}, kwargs map[string]interface{}) (rargs []interface{}, rkwargs map[string]interface{}, rerr APIError) {
	sendMsgArgs, rerr := bindSendMsgArgs(s, args)
	if rerr != nil {
		return
	}

	var ts int64
	if err := xchatLogic.Call(types.RPCXChatSendNotify, sendMsgArgs, &ts); err != nil {
		l.Warning("%s error: %s", types.RPCXChatSendNotify, err)
		rerr = newDefaultAPIError(err.Error())
		return
	}

	return []interface{}{true, ts}, nil, nil
}

func doPushUserNotify(ts int64, sendUserMsgArgs *types.SendUserMsgArgs) {
	pushUserNotify(sendUserMsgArgs.Source, &pubtypes.UserNotifyMessage{
		ToUser: sendUserMsgArgs.ToUser,
		Domain: sendUserMsgArgs.Domain,
		User:   sendUserMsgArgs.User,
		Ts:     ts,
		Msg:    sendUserMsgArgs.Msg,
	})
}

func onPubUserNotify(s *Session, args []interface{}, kwargs map[string]interface{}) {
	sendUserMsgArgs, rerr := bindSendUserMsgArgs(s, args, kwargs)
	if rerr != nil {
		return
	}

	xchatLogic.AsyncCall(types.RPCXChatSendUserNotify, sendUserMsgArgs)

	go doPushUserNotify(time.Now().Unix(), sendUserMsgArgs)
}

func sendUserNotify(s *Session, args []interface{}, kwargs map[string]interface{}) (rargs []interface{}, rkwargs map[string]interface{}, rerr APIError) {
	sendUserMsgArgs, rerr := bindSendUserMsgArgs(s, args, kwargs)
	if rerr != nil {
		return
	}

	var ts int64
	if err := xchatLogic.Call(types.RPCXChatSendUserNotify, sendUserMsgArgs, &ts); err != nil {
		l.Warning("%s error: %s", types.RPCXChatSendUserNotify, err)
		rerr = newDefaultAPIError(err.Error())
		return
	}

	go doPushUserNotify(ts, sendUserMsgArgs)

	return []interface{}{true, ts}, nil, nil
}

// 创建会话
func newChat(s *Session, args []interface{}, kwargs map[string]interface{}) (rargs []interface{}, rkwargs map[string]interface{}, rerr APIError) {
	// TODO: arguments databinding.
	chatType := args[0].(string)
	if chatType != "user" && chatType != "users" && chatType != "self" {
		rerr = InvalidChatTypeError
		return
	}

	isNsUser := false
	if x, ok := kwargs["is_ns_user"]; ok {
		isNsUser = x.(bool)
	}

	ns, _ := nsutils.DecodeNSUser(s.User)
	users := []string{s.User}
	for _, u := range args[1].([]interface{}) {
		if isNsUser {
			users = append(users, u.(string))
		} else {
			users = append(users, nsutils.EncodeNSUser(ns, u.(string)))
		}
	}
	title := args[2].(string)
	ext := ""
	if x, ok := kwargs["ext"]; ok {
		ext = x.(string)
	}

	// NOTE: 利用了ns=""的情况可以添加任何ns的用户
	xchatID, err := xchatHTTPClient.NewChat(chatType, users, title, "user", ext)
	if err != nil {
		rerr = newDefaultAPIError(err.Error())
		return
	}

	chatIdentity, err := db.ParseChatIdentity(xchatID)
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
	chatIdentity, err := db.ParseChatIdentity(args[0].(string))
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
	chatIdentity, err := db.ParseChatIdentity(args[0].(string))
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

func doSetUserChat(user string, chatID uint64, key string, value interface{}) error {
	return xchatLogic.Call(types.RPCXChatSetUserChat, &types.SetUserChatArgs{
		User:   user,
		ChatID: chatID,
		Key:    key,
		Value:  value,
	}, nil)
}

// 设置用户会话属性
func setChat(s *Session, args []interface{}, kwargs map[string]interface{}) (rargs []interface{}, rkwargs map[string]interface{}, rerr APIError) {
	chatIdentity, err := db.ParseChatIdentity(args[0].(string))
	if err != nil {
		rerr = InvalidArgumentError
		return
	}
	if chatIdentity.Type == "room" {
		// 直接成功
		return []interface{}{true}, nil, nil
	}

	chatID := chatIdentity.ID

	for key, x := range kwargs {
		switch key {
		case "session_id":
			// pass this.
			continue
		case "dnd":
			if val, ok := x.(bool); ok {
				if err := doSetUserChat(s.User, chatID, key, val); err != nil {
					rerr = newDefaultAPIError(err.Error())
					return
				}
				continue
			}
		case "cur_id":
			if val, ok := x.(float64); ok {
				if err := doSetUserChat(s.User, chatID, key, uint64(val)); err != nil {
					rerr = newDefaultAPIError(err.Error())
					return
				}
				continue
			}
		}
		rerr = InvalidArgumentError
		return
	}
	return []interface{}{true}, nil, nil
}

// 同步会话接收消息
func syncChatRecv(s *Session, args []interface{}, kwargs map[string]interface{}) (rargs []interface{}, rkwargs map[string]interface{}, rerr APIError) {
	chatIdentity, err := db.ParseChatIdentity(args[0].(string))
	if err != nil {
		rerr = InvalidArgumentError
		return
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
	chatIdentity, err := db.ParseChatIdentity(args[0].(string))
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

	if kwargs["desc"] != nil {
		desc = kwargs["desc"].(bool)
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
	for i := range msgs {
		toPushMsgs = append(toPushMsgs, NewMessageFromPubMsg(&msgs[i]))
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

	chatIdentity, err := db.ParseChatIdentity(args[1].(string))
	if err != nil {
		rerr = newDefaultAPIError(err.Error())
		return
	}
	chatID := chatIdentity.ID

	s.ExitRoom(roomID, chatID)

	return []interface{}{true}, nil, nil
}

func bindJoinExitChatArgs(s *Session, args []interface{}, kwargs map[string]interface{}) (joinExitChatArgs *types.JoinExitChatArgs, rerr APIError) {
	chatIdentity, err := db.ParseChatIdentity(args[0].(string))
	if err != nil {
		rerr = InvalidArgumentError
		return
	}
	chatID := chatIdentity.ID
	chatType := chatIdentity.Type

	isNsUser := false
	if x, ok := kwargs["is_ns_user"]; ok {
		isNsUser = x.(bool)
	}

	ns, _ := nsutils.DecodeNSUser(s.User)
	users := []string{}
	if len(args) >= 2 {
		// 邀请的人员
		if isNsUser {
			for _, u := range args[1].([]interface{}) {
				users = append(users, u.(string))
			}
		} else {
			for _, u := range args[1].([]interface{}) {
				users = append(users, nsutils.EncodeNSUser(ns, u.(string)))
			}
		}
	}

	return &types.JoinExitChatArgs{
		ChatID:   chatID,
		ChatType: chatType,
		User:     s.User,
		Users:    users,
	}, nil
}

func joinChat(s *Session, args []interface{}, kwargs map[string]interface{}) (rargs []interface{}, rkwargs map[string]interface{}, rerr APIError) {
	joinExitChatArgs, rerr := bindJoinExitChatArgs(s, args, kwargs)
	if rerr != nil {
		return
	}

	if err := xchatLogic.Call(types.RPCXChatJoinChat, joinExitChatArgs, nil); err != nil {
		rerr = newDefaultAPIError(err.Error())
		return
	}

	return []interface{}{true}, nil, nil
}

func exitChat(s *Session, args []interface{}, kwargs map[string]interface{}) (rargs []interface{}, rkwargs map[string]interface{}, rerr APIError) {
	joinExitChatArgs, rerr := bindJoinExitChatArgs(s, args, kwargs)
	if rerr != nil {
		return
	}

	if err := xchatLogic.Call(types.RPCXChatExitChat, joinExitChatArgs, nil); err != nil {
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

	chatIdentity, err := db.ParseChatIdentity(xchatID)
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
