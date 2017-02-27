package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"xim/utils/nsutils"
	"xim/xchat/logic/cache"
	"xim/xchat/logic/db"
	"xim/xchat/logic/mq"
	"xim/xchat/logic/pub"
	pubtypes "xim/xchat/logic/pub/types"
	"xim/xchat/logic/service/types"
)

// constants
const (
	NSDefault                    = ""
	NSCs                         = "cs"
	NSTest                       = "test"
	MaxRoomChatHistoryCount      = 100
	MaxUserChatExtraHistoryCount = 500
)

// errors
var (
	ErrNoPermission    = errors.New("no permission")
	ErrOperationFailed = errors.New("operation failed")
	ErrIllegalRequest  = errors.New("illegal request")
)

// msg
const (
	// domain
	XChatDomain = "_"
	// xchat msgs
	XChatDomainChatInfoUpdatedMsg    = "$CHAT_INFO_UPDATED"
	XChatDomainChatMembersUpdatedMsg = "$CHAT_MEMBERS_UPDATED"
	XChatDomainBeRemovedFromChatMsg  = "$BE_REMOVED"
)

// options
var (
	XChatDomainSendMsgOptions = &types.SendMsgOptions{
		IgnorePermCheck:     true,
		IgnoreNotifyOffline: true,
	}
)

// Ping is a test service.
func Ping(sleep int64, payload string) string {
	time.Sleep(time.Duration(sleep) * time.Millisecond)
	return "RPC:" + payload
}

// RoomExists judges whether room exists.
func RoomExists(roomID uint64) (bool, error) {
	return db.RoomExists(roomID)
}

// FetchChatMembers fetch chat's members.
func FetchChatMembers(chatID uint64) ([]db.Member, error) {
	return db.GetChatMembers(chatID)
}

// FetchChat fetch chat.
func FetchChat(chatID uint64) (*db.Chat, error) {
	return db.GetChat(chatID)
}

// FetchUserChat fetch user's chat.
func FetchUserChat(user string, chatID uint64) (*db.UserChat, error) {
	return db.GetUserChat(user, chatID)
}

// FetchUserChatList fetch user's chat list.
func FetchUserChatList(user string, onlyUnsync bool, lastMsgTs int64) ([]db.UserChat, error) {
	return db.GetUserChatList(user, onlyUnsync, lastMsgTs)
}

// SetUserChat set user's chat attribute.
func SetUserChat(user string, chatID uint64, key string, value interface{}) (time.Time, error) {
	return db.SetUserChat(user, chatID, key, value)
}

// SyncUserChatRecv sync user's chat msg recv.
func SyncUserChatRecv(user string, chatID uint64, msgID uint64) error {
	return db.SyncUserChatRecv(user, chatID, msgID)
}

// IsChatMember checks is user chat's member.
func IsChatMember(chatID uint64, user string) (bool, error) {
	return db.IsChatMember(chatID, user)
}

func msgToChatMsg(msg *db.Message, membersUpdated int64) *pubtypes.ChatMessage {
	return &pubtypes.ChatMessage{
		ChatID:         msg.ChatID,
		ChatType:       msg.ChatType,
		Domain:         msg.Domain,
		ID:             msg.ID,
		User:           msg.User,
		Ts:             msg.Ts.Unix(),
		Msg:            msg.Msg,
		MembersUpdated: membersUpdated,
	}
}

func msgsToChatMsgs(msgs []db.Message) []pubtypes.ChatMessage {
	ms := []pubtypes.ChatMessage{}
	for _, msg := range msgs {
		ms = append(ms, *msgToChatMsg(&msg, 0))
	}
	return ms
}

// FetchChatMessages fetch chat's messages between lID and rID.
func FetchChatMessages(chatID uint64, chatType string, lID, rID uint64, limit int, desc bool) ([]pubtypes.ChatMessage, error) {
	msgs, err := db.GetChatMessages(chatID, chatType, lID, rID, limit, desc)
	if err != nil {
		return nil, err
	}

	return msgsToChatMsgs(msgs), nil
}

// FetchUserChatMessages fetch chat's messages between lID and rID.
func FetchUserChatMessages(user string, chatID uint64, chatType string, lID, rID uint64, limit int, desc bool) ([]pubtypes.ChatMessage, error) {
	ns, _ := nsutils.DecodeNSUser(user)
	// 客服可以拿任意消息
	if ns != NSCs {
		if chatType == types.ChatTypeRoom {
			// 房间会话最多拿最近MaxRoomChatHistoryCount条消息
			chat, err := db.GetChatWithType(chatID, chatType)
			if err != nil {
				return nil, ErrNoPermission
			}

			if chat.MsgID > lID+MaxRoomChatHistoryCount {
				lID = chat.MsgID - MaxRoomChatHistoryCount
			}
		} else {
			userChat, err := db.GetUserChatWithType(user, chatID, chatType)
			if err != nil {
				return nil, ErrNoPermission
			}
			// 普通会话可以获取额外MaxUserChatExtraHistoryCount的历史消息
			if userChat.JoinMsgID > lID+MaxUserChatExtraHistoryCount {
				lID = userChat.JoinMsgID - MaxUserChatExtraHistoryCount
			}
		}
	}
	return FetchChatMessages(chatID, chatType, lID, rID, limit, desc)
}

// FetchChatMessagesByIDs fetch chat's messages by ids.
func FetchChatMessagesByIDs(chatID uint64, chatType string, msgIDs []uint64) ([]pubtypes.ChatMessage, error) {
	msgs, err := db.GetChatMessagesByIDs(chatID, chatType, msgIDs)
	if err != nil {
		return nil, err
	}

	return msgsToChatMsgs(msgs), nil
}

// default users.
const (
	CSUser = "cs:cs"
)

// SendUserNotify sends user notify.
func SendUserNotify(src *pubtypes.MsgSource, toUser, domain, user, msg string, options *types.SendMsgOptions) (int64, error) {
	if !(options != nil && options.IgnorePermCheck) {
		if user != toUser {
			t, err := db.IsHaveUserChat(user, toUser)
			if err != nil {
				//return 0, err
				return 0, ErrNoPermission
			}
			if !t {
				//return 0, fmt.Errorf("no user chat between %s and %s", user, toUser)
				return 0, ErrNoPermission
			}
		}
	}

	// TODO: 解决呼叫信息的通知问题
	ok, err := cache.IsUserOnline(toUser)
	if err != nil {
		return 0, err
	}
	if !ok {
		return 0, fmt.Errorf("user is offline")
	}

	ts := time.Now()
	m := pubtypes.UserNotifyMessage{
		ToUser: toUser,
		Domain: domain,
		User:   user,
		Ts:     ts.Unix(),
		Msg:    msg,
	}

	// FIXME: goroutine pool?
	go pub.PublishMessage(&pubtypes.XMessage{
		Source: src,
		Msg:    m,
	})

	return m.Ts, nil
}

func checkSendPermissions(chatID uint64, chatType, user string, options *types.SendMsgOptions) (int64, error) {
	var membersUpdated int64
	if options != nil && options.IgnorePermCheck {
		chat, err2 := db.GetChatWithType(chatID, chatType)
		if err2 != nil {
			//return 0, fmt.Errorf("no permission: %s", err2.Error())
			return 0, ErrNoPermission
		}
		membersUpdated = chat.MembersUpdated.Unix()
	} else {
		userChat, err := db.GetUserChatWithType(user, chatID, chatType)
		if err != nil {
			// FIXME: 目前房间和客服可以随意发消息
			if chatType != types.ChatTypeRoom && user != CSUser {
				//return 0, fmt.Errorf("no permission: %s", err.Error())
				return 0, ErrNoPermission
			}

			chat, err2 := db.GetChatWithType(chatID, chatType)
			if err2 != nil {
				//return 0, fmt.Errorf("no permission: %s", err2.Error())
				return 0, ErrNoPermission
			}
			membersUpdated = chat.MembersUpdated.Unix()
		} else {
			membersUpdated = userChat.MembersUpdated.Unix()
		}
	}
	return membersUpdated, nil
}

// SendChatMsg sends chat message.
func SendChatMsg(src *pubtypes.MsgSource, chatID uint64, chatType, domain, user, msg string,
	forceNotifyUsers map[string]struct{},
	options *types.SendMsgOptions) (*pubtypes.ChatMessage, error) {
	membersUpdated, err := checkSendPermissions(chatID, chatType, user, options)
	if err != nil {
		return nil, err
	}

	message, err := db.NewMsg(chatID, chatType, domain, user, msg)
	if err != nil {
		return nil, err
	}

	// NOTE: members_updated 为会话更新时间, 用来判断是否更新members缓存
	m := *msgToChatMsg(message, membersUpdated)

	// FIXME: implement custom service.
	if m.ChatType == "cs" {
		if !strings.HasPrefix(m.User, "cs:") {
			var ms []db.Member
			// 判断是否有客服人员接入
			ms, err = db.GetChatMembers(chatID)
			if err == nil && len(ms) == 1 && ms[0].User == m.User {
				// 还没接入
				// TODO: remove this if custom service has implemented.
				//SendChatMsg(m.ChatID, m.ChatType, CSUser, fmt.Sprintf("{\"text\":\"%s\",\"messageType\":0}",
				//	"您好，由于现在咨询人数较多，可能无法及时回复您，您可以先完整描述您的问题，我们会尽快为您解决~"), nil)
				go publishCSRequest(m.User, m.ChatID, types.MsgKindChat, m.ID, m.Msg, message.Ts)
			}
		}
	}

	// FIXME: goroutine pool?
	go pub.PublishMessage(&pubtypes.XMessage{
		Source: src,
		Msg:    m,
	})

	if options == nil || !options.IgnoreNotifyOffline {
		go notifyOfflineUsers(message.User, chatID, types.MsgKindChat, chatType, domain, message.Msg, message.Ts, forceNotifyUsers)
	}

	return &m, err
}

// SendChatNotifyMsg sends chat notify message.
func SendChatNotifyMsg(src *pubtypes.MsgSource, chatID uint64, chatType, domain, user, msg string,
	forceNotifyUsers map[string]struct{},
	options *types.SendMsgOptions) (int64, error) {
	membersUpdated, err := checkSendPermissions(chatID, chatType, user, options)
	if err != nil {
		return 0, err
	}

	ts := time.Now()
	m := pubtypes.ChatNotifyMessage{
		ChatID:         chatID,
		ChatType:       chatType,
		Domain:         domain,
		User:           user,
		Ts:             ts.Unix(),
		Msg:            msg,
		MembersUpdated: membersUpdated,
	}
	// FIXME: goroutine pool?
	go pub.PublishMessage(&pubtypes.XMessage{
		Source: src,
		Msg:    m,
	})
	go notifyOfflineUsers(m.User, chatID, types.MsgKindChatNotify, chatType, domain, m.Msg, ts, forceNotifyUsers)

	return ts.Unix(), nil
}

// PubUserStatus publish user's status msg.
func PubUserStatus(instanceID, sessionID uint64, user string, status string) error {
	l.Debug("instance:%d, session:%d, user:%s, status:%s", instanceID, sessionID, user, status)
	// 记录用户在线状态
	if err := UpdateUserStatus(instanceID, sessionID, user, status); err != nil {
		return err
	}

	t := time.Now()
	// 发送上下线日志
	publishUserStatus(user, status, t)

	return nil
}

// PubUserInfo publish user's status info.
func PubUserInfo(instanceID, sessionID uint64, user string, status string, info string) error {
	l.Debug("instance:%d, session:%d, user:%s, status:%s, info:%s", instanceID, sessionID, user, status, info)

	t := time.Now()
	// 发送用户信息
	if info != "*" && info != "" {
		msg := make(map[string]string)
		msg["user"] = user
		msg["type"] = status
		msg["info"] = info
		msg["ts"] = strconv.FormatInt(t.Unix(), 10)

		b, err := json.Marshal(&msg)
		if err != nil {
			return err
		}

		return mq.Publish(mq.XChatLogsTopic, string(b))
	}
	return nil
}

// FetchNewRoomChats fetch room chats' ids.
func FetchNewRoomChats(roomID uint64, chatIDs []uint64) ([]db.RoomChat, error) {
	return db.GetOrCreateNewRoomChats(roomID, chatIDs)
}

// JoinChat add user to chat.
func JoinChat(chatID uint64, chatType string, user string, users []string) (err error) {
	var limit int
	ns, _ := nsutils.DecodeNSUser(user)
	// 加入规则
	// 只有cs和users会话可以加入
	// cs只可以加入cs会话
	switch chatType {
	case types.ChatTypeCS:
		// 只能加入自己
		if ns != NSCs || len(users) > 0 {
			return ErrNoPermission
		}
		limit = 1
		users = []string{user}
	case types.ChatTypeUsers:
		if ns == NSCs {
			return ErrNoPermission
		}
		// 必须是会话成员
		if _, err = db.GetUserChatWithType(user, chatID, chatType); err != nil {
			return ErrNoPermission
		}
		// 只能邀请
		if len(users) == 0 {
			return nil
		}
		// 限制users会话成员99人
		limit = 99 - len(users)
	default:
		return ErrNoPermission
	}

	if err = db.AddChatMembers(chatID, users, limit); err == nil {
		if chatType == types.ChatTypeUsers {
			SendChatMsg(nil, chatID, chatType, XChatDomain, user, XChatDomainChatMembersUpdatedMsg, nil, XChatDomainSendMsgOptions)
		}
	} else {
		err = ErrOperationFailed
	}
	return
}

// ExitChat remove user from chat.
func ExitChat(chatID uint64, chatType string, user string, users []string) (err error) {
	ns, _ := nsutils.DecodeNSUser(user)
	// 退出规则
	// 只有cs和users会话可以退出
	// 必须是会话成员
	if _, err = db.GetUserChatWithType(user, chatID, chatType); err != nil {
		return ErrNoPermission
	}

	// 删除的用户
	var (
		c                     *db.Chat
		deletedUsersSelfChats []*db.Chat
	)

	switch chatType {
	case types.ChatTypeCS:
		if ns != NSCs {
			return ErrIllegalRequest
		}
		users = []string{user}
	case types.ChatTypeUsers:
		// 当有users的时候则是请出成员, 否则为自己离开
		if len(users) == 0 {
			users = []string{user}
		}
		// 获取所有用户的self chats.
		for _, u := range users {
			c, err = db.GetOrCreateSelfChat(u)
			if err != nil {
				return ErrOperationFailed
			}
			deletedUsersSelfChats = append(deletedUsersSelfChats, c)
		}
	default:
		return ErrIllegalRequest
	}
	if err = db.RemoveChatMembers(chatID, users); err == nil {
		if chatType == types.ChatTypeUsers {
			// FIXME && TODO: 要保证一定发送成功
			// 通知成员变化
			SendChatMsg(nil, chatID, chatType, XChatDomain, user, XChatDomainChatMembersUpdatedMsg, nil, XChatDomainSendMsgOptions)
			// 通知被删除的用户自己被从该会话删除
			beRemovedMsg := fmt.Sprintf(`%s,%s.%d`, XChatDomainBeRemovedFromChatMsg, chatType, chatID)
			for _, c := range deletedUsersSelfChats {
				SendChatMsg(nil, c.ID, c.Type, XChatDomain, c.Owner.String, beRemovedMsg, nil, XChatDomainSendMsgOptions)
			}
		}
	}
	return
}

// SetChatTitle set chat's title.
func SetChatTitle(user string, chatID uint64, chatType string, title string) (err error) {
	if err = db.SetUserChatTitle(user, chatID, chatType, title); err == nil {
		SendChatMsg(nil, chatID, chatType, XChatDomain, user, XChatDomainChatInfoUpdatedMsg, nil, XChatDomainSendMsgOptions)
	} else {
		err = ErrNoPermission
	}
	return
}
