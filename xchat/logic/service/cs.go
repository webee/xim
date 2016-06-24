package service

import (
	"encoding/json"
	"time"
	"xim/xchat/logic/mq"
)

type csReq struct {
	User   string    `json:"user"`
	ChatID uint64    `json:"chat_id"`
	Kind   string    `json:"kind"`
	MsgID  uint64    `json:"msg_id"`
	Msg    string    `json:"msg"`
	Ts     time.Time `json:"ts"`
}

func publishCSRequest(user string, chatID uint64, kind string, msgID uint64, msg string, ts time.Time) {
	m := csReq{
		User:   user,
		ChatID: chatID,
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
