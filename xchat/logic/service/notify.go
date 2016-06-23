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
	MsgID    uint64    `json:"msg_id"`
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

func notifyOfflineUsers(msg *db.Message) {
	if !offlineNotifyEnabledChatTypes[msg.ChatType] {
		return
	}

	m := offlineMsg{
		From:     msg.User,
		ChatID:   msg.ChatID,
		ChatType: msg.ChatType,
		MsgID:    msg.ID,
		Msg:      msg.Msg,
		Ts:       msg.Ts,
	}

	members, err := db.GetChatMembers(msg.ChatID)
	if err != nil {
		l.Warning("get chat members error: %s", err.Error())
		return
	}

	users := []string{}
	for _, member := range members {
		if member.User != msg.User {
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
