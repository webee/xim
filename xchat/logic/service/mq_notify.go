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
	Domain   string    `json:"domain"`
	Msg      string    `json:"msg"`
	Ts       time.Time `json:"ts"`
}

var (
	offlineNotifyEnabledChatTypes = map[string]bool{
		types.ChatTypeSelf:  true,
		types.ChatTypeUser:  true,
		types.ChatTypeUsers: true,
		types.ChatTypeGroup: true,
		types.ChatTypeCS:    true,
	}
)

func notifyOfflineUsers(from string, chatID uint64, kind, chatType, domain, msg string,
	ts time.Time, forceNotifyUsers map[string]struct{}) {
	if !offlineNotifyEnabledChatTypes[chatType] {
		return
	}

	m := offlineMsg{
		From:     from,
		ChatID:   chatID,
		ChatType: chatType,
		Kind:     kind,
		Domain:   domain,
		Msg:      msg,
		Ts:       ts,
	}

	members, err := db.GetFullChatMembers(chatID)
	if err != nil {
		l.Warning("get chat members error: %s", err.Error())
		return
	}

	users := []string{}
	for _, member := range members {
		// 不发通知给自己，免打扰，已退出的
		if member.User == from || member.Dnd || member.IsExited {
			if _, ok := forceNotifyUsers[member.User]; !ok {
				continue
			}
		}
		users = append(users, member.User)
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
		l.Debug("notify %s, %+v", user, m)
		b, err := json.Marshal(&m)
		if err != nil {
			l.Warning("json encoding error: %s", err.Error())
			return
		}

		mq.Publish(mq.XChatUserMsgsTopic, string(b))
	}
}
