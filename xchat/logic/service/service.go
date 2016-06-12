package service

import (
	"fmt"
	"time"
	"xim/xchat/logic/db"
	"xim/xchat/logic/pub"
	pubtypes "xim/xchat/logic/pub/types"
)

// Echo send msg back.
func Echo(s string) string {
	return s
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
func FetchChatMessages(chatID uint64, lID, rID uint64, limit int, desc bool) ([]pubtypes.ChatMessage, error) {
	msgs, err := db.GetChatMessages(chatID, lID, rID, limit, desc)
	if err != nil {
		return nil, err
	}
	ms := []pubtypes.ChatMessage{}
	for _, msg := range msgs {
		ms = append(ms, pubtypes.ChatMessage{
			ChatID: msg.ChatID,
			ID:     msg.ID,
			User:   msg.User,
			Ts:     msg.Ts.Unix(),
			Msg:    msg.Msg,
		})
	}
	return ms, nil
}

// SendChatMsg sends chat message.
func SendChatMsg(chatID uint64, user string, msg string) (*pubtypes.ChatMessage, error) {
	message, err := db.NewMsg(chatID, user, msg)
	if err != nil {
		return nil, err
	}

	// publish
	m := pubtypes.ChatMessage{
		ChatID: message.ChatID,
		ID:     message.ID,
		User:   message.User,
		Ts:     message.Ts.Unix(),
		Msg:    message.Msg,
	}
	// FIXME: goroutine pool?
	go pub.PublishMessage(&pubtypes.XMessage{
		Msg: m,
	})
	return &m, err
}

// SendChatNotifyMsg sends chat notify message.
func SendChatNotifyMsg(chatID uint64, user string, msg string) error {
	ok, err := db.IsChatMember(chatID, user)
	if err != nil {
		return err
	}

	if !ok {
		return fmt.Errorf("no permission")
	}

	m := pubtypes.ChatNotifyMessage{
		ChatID: chatID,
		User:   user,
		Ts:     time.Now().Unix(),
		Msg:    msg,
	}
	go pub.PublishMessage(&pubtypes.XMessage{
		Msg: m,
	})
	return nil
}
