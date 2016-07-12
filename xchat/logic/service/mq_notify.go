package service

import (
	"encoding/json"
	"time"
	"xim/xchat/logic/cache"
	"xim/xchat/logic/db"
	"xim/xchat/logic/mq"
	"xim/xchat/logic/service/types"
)

type offlineMsg struct {
	User     string    `json:"user"`
	From     string    `json:"from"`
	ChatID   uint64    `json:"chat_id"`
	ChatType string    `json:"chat_type"`
	Kind     string    `json:"kind"`
	Msg      string    `json:"msg"`
	Ts       time.Time `json:"ts"`
}

var (
	offlineNotifyEnabledChatTypes = map[string]bool{
		types.ChatTypeUser:  true,
		types.ChatTypeGroup: true,
		types.ChatTypeCS:    true,
	}
)

func notifyOfflineUsers(from string, chatID uint64, kind, chatType, msg string, ts time.Time) {
	if !offlineNotifyEnabledChatTypes[chatType] {
		return
	}

	m := offlineMsg{
		From:     from,
		ChatID:   chatID,
		Kind:     kind,
		ChatType: chatType,
		Msg:      msg,
		Ts:       ts,
	}

	members, err := db.GetChatMembers(chatID)
	if err != nil {
		l.Warning("get chat members error: %s", err.Error())
		return
	}

	users := []string{}
	for _, member := range members {
		if member.User != from {
			users = append(users, member.User)
		}
	}
	if len(users) == 0 {
		return
	}

	offlineUsers, err := cache.GetOfflineUsers(users...)
	if err != nil {
		l.Warning("get offline users error: %s", err.Error())
		return
	}

	for _, user := range offlineUsers {
		m.User = user
		b, err := json.Marshal(&m)
		if err != nil {
			l.Warning("json encoding error: %s", err.Error())
			return
		}

		mq.Publish(mq.XChatUserMsgsTopic, string(b))
	}
}
