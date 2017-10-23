package service

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
	"xim/xchat/logic/db"
	"xim/xchat/logic/service/types"
)

var (
	client = &http.Client{}
)

func notifyChatMessage(appID string, msg *db.Message) {
	app := getAppInfo(appID)
	if app == nil || !app.MsgNotifyURL.Valid || app.MsgNotifyURL.String == "" {
		return
	}
	chatIdentity := db.ChatIdentity{
		ID:   msg.ChatID,
		Type: msg.ChatType,
	}
	params := make(map[string]interface{})
	params["kind"] = types.MsgKindChat
	params["chat_id"] = chatIdentity.String()
	params["uid"] = msg.User
	params["id"] = msg.ID
	params["msg"] = msg.Msg
	params["ts"] = msg.Ts
	params["domain"] = msg.Domain

	go doNotify(params, app.MsgNotifyURL.String)
}

func notifyChatNotifyMessage(appID string, chatID uint64, chatType, user, msg string, ts time.Time, domain string) {
	app := getAppInfo(appID)
	if app == nil || !app.MsgNotifyURL.Valid || app.MsgNotifyURL.String == "" {
		return
	}
	chatIdentity := db.ChatIdentity{
		ID:   chatID,
		Type: chatType,
	}
	params := make(map[string]interface{})
	params["kind"] = types.MsgKindChatNotify
	params["chat_id"] = chatIdentity.String()
	params["uid"] = user
	params["msg"] = msg
	params["ts"] = ts
	params["domain"] = domain

	go doNotify(params, app.MsgNotifyURL.String)
}

func doNotify(params map[string]interface{}, url string) {
	b, _ := json.Marshal(params)
	req, err := http.NewRequest("POST", url, bytes.NewReader(b))
	if err != nil {
		l.Warning("notify %s: %v", url, err)
		return
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		l.Warning("notify %s: %v", url, err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		l.Warning("notify %s failed: status code %d", url, resp.StatusCode)
		return
	}
}
