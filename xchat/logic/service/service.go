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
	pub.PublishMessage(m)

	return m, err
}
