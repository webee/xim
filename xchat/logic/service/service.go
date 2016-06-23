package service

import (
	"encoding/json"
	"fmt"
	"time"
	"xim/xchat/logic/db"
	"xim/xchat/logic/mq"
	"xim/xchat/logic/pub"
	pubtypes "xim/xchat/logic/pub/types"
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

// default users.
const (
	CSUser = "cs:cs"
)

// SendChatMsg sends chat message.
func SendChatMsg(chatID uint64, user string, msg string) (*pubtypes.ChatMessage, error) {
	var updated int64
	var chatType string
	userChat, err := db.GetUserChat(user, chatID)
	if err != nil {
		chat, err2 := db.GetChat(chatID)
		if err2 != nil {
			return nil, fmt.Errorf("no permission: %s", err2.Error())
		}
		// FIXME: implement custom service.
		if chat.Type != "room" && user != CSUser {
			return nil, fmt.Errorf("no permission: %s", err.Error())
		}

		// is room chat.
		chatType = chat.Type
	} else {
		updated = userChat.Updated.Unix()
		chatType = userChat.Type
	}

	message, err := db.NewMsg(chatID, chatType, user, msg)
	if err != nil {
		return nil, err
	}

	m := pubtypes.ChatMessage{
		ChatID:   message.ChatID,
		ChatType: message.ChatType,
		ID:       message.ID,
		User:     message.User,
		Ts:       message.Ts.Unix(),
		Msg:      message.Msg,
		Updated:  updated,
	}
	// FIXME: implement custom service.
	if m.ChatType == "cs" {
		if m.User != CSUser {
			SendChatMsg(m.ChatID, CSUser, fmt.Sprintf("{\"text\":\"%s\",\"messageType\":0}",
				"您好，由于现在咨询人数较多，可能无法及时回复您，您可以先完整描述您的问题，我们会尽快为您解决~"))
		}
	}

	// FIXME: goroutine pool?
	go pub.PublishMessage(&pubtypes.XMessage{
		Msg: m,
	})
	go notifyOfflineUsers(message)

	return &m, err
}

// SendChatNotifyMsg sends chat notify message.
func SendChatNotifyMsg(chatID uint64, user string, msg string) error {
	var updated int64
	var chatType string
	userChat, err := db.GetUserChat(user, chatID)
	if err != nil {
		chat, err2 := db.GetChat(chatID)
		if err2 != nil {
			return fmt.Errorf("no permission: %s", err2.Error())
		}
		if chat.Type != "room" {
			return fmt.Errorf("no permission: %s", err.Error())
		}
		// is room chat.
		chatType = chat.Type
	} else {
		updated = userChat.Updated.Unix()
		chatType = userChat.Type
	}

	m := pubtypes.ChatNotifyMessage{
		ChatID:   chatID,
		ChatType: chatType,
		User:     user,
		Ts:       time.Now().Unix(),
		Msg:      msg,
		Updated:  updated,
	}
	// FIXME: goroutine pool?
	go pub.PublishMessage(&pubtypes.XMessage{
		Msg: m,
	})
	return nil
}

// PubUserStatus publish user's status msg.
func PubUserStatus(instanceID, sessionID uint64, user string, status string, info string) error {
	l.Debug("instance:%d, session:%d, user:%s, status:%s, info:%s", instanceID, sessionID, user, status, info)
	// 记录用户在线状态
	UpdateUserStatus(instanceID, sessionID, user, status)

	// 发送上下线日志
	if info != "" {
		msg := make(map[string]string)
		msg["user"] = user
		msg["type"] = status
		msg["info"] = info

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
