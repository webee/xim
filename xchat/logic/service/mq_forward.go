package service

import (
	"encoding/json"
	"time"
	"xim/xchat/logic/db"
	"xim/xchat/logic/mq"
	"xim/xchat/logic/service/types"
)

type forwardedMsg struct {
	Kind   string    `json:"kind"`
	ChatID string    `json:"chat_id"`
	UID    string    `json:"uid"`
	ID     uint64    `json:"id,omitempty"`
	Msg    string    `json:"msg"`
	Ts     time.Time `json:"ts"`
	Domain string    `json:"domain"`
}

func forwardChatMessage(mqTopic string, msg *db.Message) {
	chatIdentity := db.ChatIdentity{
		ID:   msg.ChatID,
		Type: msg.ChatType,
	}
	m := forwardedMsg{
		Kind:   types.MsgKindChat,
		ChatID: chatIdentity.String(),
		UID:    msg.User,
		ID:     msg.ID,
		Msg:    msg.Msg,
		Ts:     msg.Ts,
		Domain: msg.Domain,
	}

	b, err := json.Marshal(&m)
	if err != nil {
		l.Warning("json encoding error: %s", err.Error())
		return
	}

	mq.PublishBytesWithKey(mqTopic, m.ChatID, b)
}

func forwardChatNotifyMessage(mqTopic string, chatID uint64, chatType, user, msg string, ts time.Time, domain string) {
	chatIdentity := db.ChatIdentity{
		ID:   chatID,
		Type: chatType,
	}
	m := forwardedMsg{
		Kind:   types.MsgKindChatNotify,
		ChatID: chatIdentity.String(),
		UID:    user,
		Msg:    msg,
		Ts:     ts,
		Domain: domain,
	}

	b, err := json.Marshal(&m)
	if err != nil {
		l.Warning("json encoding error: %s", err.Error())
		return
	}

	mq.PublishBytesWithKey(mqTopic, m.ChatID, b)
}
