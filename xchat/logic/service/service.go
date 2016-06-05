package service

import (
	"xim/xchat/logic/db"
	"xim/xchat/logic/pub"
	"xim/xchat/logic/pub/types"
)

// Echo send msg back.
func Echo(s string) string {
	return s
}

// FetchChatMembers fetch chat's members.
func FetchChatMembers(chatID uint64) ([]db.Member, error) {
	return db.GetChatMembers(chatID)
}

// FetchChatMessages fetch chat's messages between sID and eID.
func FetchChatMessages(chatID uint64, sID, eID uint64) ([]types.Message, error) {
	msgs, err := db.GetChatMessages(chatID, sID, eID)
	if err != nil {
		return nil, err
	}
	ms := []types.Message{}
	for _, msg := range msgs {
		ms = append(ms, types.Message{
			ChatID: msg.ChatID,
			MsgID:  msg.MsgID,
			User:   msg.User,
			Ts:     msg.Ts.Unix(),
			Msg:    msg.Msg,
		})
	}
	return ms, nil
}

// SendMsg sends message.
func SendMsg(chatID uint64, user string, msg string) (*types.Message, error) {
	message, err := db.NewMsg(chatID, user, msg)
	if err != nil {
		return nil, err
	}

	// publish
	m := &types.Message{
		ChatID: message.ChatID,
		MsgID:  message.MsgID,
		User:   message.User,
		Ts:     message.Ts.Unix(),
		Msg:    message.Msg,
	}
	// FIXME: goroutine pool?
	go pub.PublishMessage(m)

	return m, err
}
