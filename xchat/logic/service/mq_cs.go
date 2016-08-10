package service

import (
	"encoding/json"
	"time"
	"xim/xchat/logic/db"
	"xim/xchat/logic/mq"
	"xim/xchat/logic/service/types"
)

type csReq struct {
	User   string    `json:"user"`
	ChatID string    `json:"chat_id"`
	Kind   string    `json:"kind"`
	MsgID  uint64    `json:"msg_id"`
	Msg    string    `json:"msg"`
	Ts     time.Time `json:"ts"`
}

func publishCSRequest(user string, chatID uint64, kind string, msgID uint64, msg string, ts time.Time) {
	chatIdentity := db.ChatIdentity{
		ID:   chatID,
		Type: types.ChatTypeCS,
	}
	m := csReq{
		User:   user,
		ChatID: chatIdentity.String(),
		Kind:   kind,
		MsgID:  msgID,
		Msg:    msg,
		Ts:     ts,
	}

	b, err := json.Marshal(&m)
	if err != nil {
		l.Warning("json encoding error: %s", err.Error())
		return
	}

	mq.Publish(mq.XChatCSReqs, string(b))
}
