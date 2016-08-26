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
	NSDefault = ""
	NSCs      = "cs"
	NSTest    = "test"
)

// errors
var (
	ErrNoPermission     = errors.New("no permission")
	ErrIllegalOperation = errors.New("illegal operation")
)

// Ping is a test service.
func Ping(sleep int64, payload string) string {
	time.Sleep(time.Duration(sleep) * time.Millisecond)
	return "RPC:" + payload
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
func FetchUserChatList(user string, onlyUnsync bool) ([]db.UserChat, error) {
	return db.GetUserChatList(user, onlyUnsync)
}

// SetUserChat set user's chat attribute.
func SetUserChat(user string, chatID uint64, key string, value interface{}) error {
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

// FetchChatMessages fetch chat's messages between lID and rID.
func FetchChatMessages(chatID uint64, chatType string, lID, rID uint64, limit int, desc bool) ([]pubtypes.ChatMessage, error) {
	msgs, err := db.GetChatMessages(chatID, chatType, lID, rID, limit, desc)
	if err != nil {
		return nil, err
	}
	ms := []pubtypes.ChatMessage{}
	for _, msg := range msgs {
		ms = append(ms, pubtypes.ChatMessage{
			ChatID:   msg.ChatID,
			ChatType: msg.ChatType,
			ID:       msg.ID,
			User:     msg.User,
			Ts:       msg.Ts.Unix(),
			Msg:      msg.Msg,
		})
	}
	return ms, nil
}

// FetchChatMessagesByIDs fetch chat's messages by ids.
func FetchChatMessagesByIDs(chatID uint64, chatType string, msgIDs []uint64) ([]pubtypes.ChatMessage, error) {
	msgs, err := db.GetChatMessagesByIDs(chatID, chatType, msgIDs)
	if err != nil {
		return nil, err
	}
	ms := []pubtypes.ChatMessage{}
	for _, msg := range msgs {
		ms = append(ms, pubtypes.ChatMessage{
			ChatID:   msg.ChatID,
			ChatType: msg.ChatType,
			ID:       msg.ID,
			User:     msg.User,
			Ts:       msg.Ts.Unix(),
			Msg:      msg.Msg,
		})
	}
	return ms, nil
}

// default users.
const (
	CSUser = "cs:cs"
)

// SendUserNotify sends user notify.
func SendUserNotify(src *pubtypes.MsgSource, toUser, domain, user, msg string) (int64, error) {
	ts := time.Now()
	m := pubtypes.UserNotifyMessage{
		ToUser: toUser,
		Domain: domain,
		User:   user,
		Ts:     ts.Unix(),
		Msg:    msg,
	}

	// TODO: 解决呼叫信息的通知问题
	if ok, err := cache.IsUserOnline(toUser); !ok {
		return 0, err
	}

	// FIXME: goroutine pool?
	go pub.PublishMessage(&pubtypes.XMessage{
		Source: src,
		Msg:    m,
	})

	return m.Ts, nil
}

// SendChatMsg sends chat message.
func SendChatMsg(src *pubtypes.MsgSource, chatID uint64, chatType string, domain string, user string, msg string) (*pubtypes.ChatMessage, error) {
	var updated int64
	userChat, err := db.GetUserChatWithType(user, chatID, chatType)
	if err != nil {
		chat, err2 := db.GetChatWithType(chatID, chatType)
		if err2 != nil {
			return nil, fmt.Errorf("no permission: %s", err2.Error())
		}
		// FIXME: implement custom service.
		if chat.Type != "room" && user != CSUser {
			return nil, fmt.Errorf("no permission: %s", err.Error())
		}
	} else {
		updated = userChat.Updated.Unix()
	}

	message, err := db.NewMsg(chatID, chatType, domain, user, msg)
	if err != nil {
		return nil, err
	}

	// NOTE: updated 为会话更新时间, 用来判断是否更新members缓存
	m := pubtypes.ChatMessage{
		ChatID:   message.ChatID,
		ChatType: message.ChatType,
		Domain:   message.Domain,
		ID:       message.ID,
		User:     message.User,
		Ts:       message.Ts.Unix(),
		Msg:      message.Msg,
		Updated:  updated,
	}
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
				//	"您好，由于现在咨询人数较多，可能无法及时回复您，您可以先完整描述您的问题，我们会尽快为您解决~"))
				go publishCSRequest(m.User, m.ChatID, types.MsgKindChat, m.ID, m.Msg, message.Ts)
			}
		}
	}

	// FIXME: goroutine pool?
	go pub.PublishMessage(&pubtypes.XMessage{
		Source: src,
		Msg:    m,
	})
	go notifyOfflineUsers(message.User, chatID, types.MsgKindChat, chatType, domain, message.Msg, message.Ts)

	return &m, err
}

// SendChatNotifyMsg sends chat notify message.
func SendChatNotifyMsg(src *pubtypes.MsgSource, chatID uint64, chatType string, domain string, user string, msg string) (int64, error) {
	var updated int64
	userChat, err := db.GetUserChatWithType(user, chatID, chatType)
	if err != nil {
		chat, err2 := db.GetChatWithType(chatID, chatType)
		if err2 != nil {
			return 0, fmt.Errorf("no permission: %s", err2.Error())
		}
		if chat.Type != "room" {
			return 0, fmt.Errorf("no permission: %s", err.Error())
		}
	} else {
		updated = userChat.Updated.Unix()
	}

	ts := time.Now()
	m := pubtypes.ChatNotifyMessage{
		ChatID:   chatID,
		ChatType: chatType,
		Domain:   domain,
		User:     user,
		Ts:       ts.Unix(),
		Msg:      msg,
		Updated:  updated,
	}
	// FIXME: goroutine pool?
	go pub.PublishMessage(&pubtypes.XMessage{
		Source: src,
		Msg:    m,
	})
	go notifyOfflineUsers(m.User, chatID, types.MsgKindChatNotify, chatType, domain, m.Msg, ts)

	return ts.Unix(), nil
}

// PubUserStatus publish user's status msg.
func PubUserStatus(instanceID, sessionID uint64, user string, status string, info string) error {
	l.Debug("instance:%d, session:%d, user:%s, status:%s, info:%s", instanceID, sessionID, user, status, info)
	// 记录用户在线状态
	if err := UpdateUserStatus(instanceID, sessionID, user, status); err != nil {
		return err
	}

	t := time.Now()
	publishUserStatus(user, status, t)

	// 发送上下线日志
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

// FetchNewRoomChatIDs fetch room chats' ids.
func FetchNewRoomChatIDs(roomID uint64, chatIDs []uint64) ([]uint64, error) {
	return db.GetOrCreateNewRoomChatIDs(roomID, chatIDs)
}

// JoinChat add user to chat.
func JoinChat(chatID uint64, chatType string, user string, users []string) error {
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
		if _, err := db.GetUserChatWithType(user, chatID, chatType); err != nil {
			return ErrNoPermission
		}
		// 只能邀请
		if len(users) == 0 {
			return nil
		}
	default:
		return ErrNoPermission
	}

	return db.AddChatMembers(chatID, users, limit)
}

// ExitChat remove user from chat.
func ExitChat(chatID uint64, chatType string, user string, users []string) error {
	ns, _ := nsutils.DecodeNSUser(user)
	// 退出规则
	// 只有cs和users会话可以退出
	switch chatType {
	case types.ChatTypeCS:
		if ns != NSCs {
			return ErrIllegalOperation
		}
		users = []string{user}
	case types.ChatTypeUsers:
		// 必须是会话成员
		if _, err := db.GetUserChatWithType(user, chatID, chatType); err != nil {
			return ErrNoPermission
		}
		// 当有users的时候则是请出成员, 否则为自己离开
		if len(users) == 0 {
			users = []string{user}
		}
	default:
		return ErrIllegalOperation
	}
	return db.RemoveChatMembers(chatID, users)
}
