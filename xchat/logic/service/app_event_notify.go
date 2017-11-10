package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"time"
	"xim/utils/nsutils"
	"xim/xchat/logic/db"
	"xim/xchat/logic/service/types"
)

var (
	errInvalidURL = errors.New("invalid url")

	client = &http.Client{}
)

// UserStatus represents user's status
type UserStatus struct {
	User   string `json:"user"`
	Status string `json:"status"`
}

func getAppEventNotifyURL(appID string) (url string, err error) {
	app := getAppInfo(appID)
	if app == nil || !app.EventNotifyURL.Valid || app.EventNotifyURL.String == "" {
		err = errInvalidURL
		return
	}
	url = app.EventNotifyURL.String
	return
}

func appNotifyUserStatus(userStatuses []UserStatus) {
	appUserStatuses := make(map[string][]UserStatus)
	for _, userStatus := range userStatuses {
		appID, _ := nsutils.DecodeNSUser(userStatus.User)
		statuses := appUserStatuses[appID]
		statuses = append(statuses, userStatus)
		appUserStatuses[appID] = statuses
	}

	for appID, userStatuses := range appUserStatuses {
		eventNotifyURL, err := getAppEventNotifyURL(appID)
		if err != nil {
			continue
		}
		data := make(map[string]interface{})
		data["statuses"] = userStatuses
		// FIXME: 考虑分次发送
		go doNotify("user_status", data, eventNotifyURL)
	}
}

// notifyChatMessage notity msg: chat
func appNotifyChatMessage(appID string, msg *db.Message) {
	notifyMessage(appID, types.MsgKindChat, msg.ChatID, msg.ChatType, msg.User, msg.Msg, msg.ID, msg.Ts, msg.Domain)
}

// notifyChatNotifyMessage notify msg: chat_notify
func appNotifyChatNotifyMessage(appID string, chatID uint64, chatType, user, msg string, ts time.Time, domain string) {
	notifyMessage(appID, types.MsgKindChatNotify, chatID, chatType, user, msg, 0, ts, domain)
}

func notifyMessage(appID, kind string, chatID uint64, chatType, user, msg string, id uint64, ts time.Time, domain string) {
	eventNotifyURL, err := getAppEventNotifyURL(appID)
	if err != nil {
		return
	}

	chatIdentity := db.ChatIdentity{
		ID:   chatID,
		Type: chatType,
	}
	data := make(map[string]interface{})
	data["kind"] = kind
	data["chat_id"] = chatIdentity.String()
	data["user"] = user
	if id > 0 {
		data["id"] = id
	}
	data["msg"] = msg
	data["ts"] = ts.Unix()
	data["domain"] = domain

	go doNotify("msg", data, eventNotifyURL)
}

func doNotify(event string, data map[string]interface{}, url string) {
	l.Debug("notify %s: %+v", event, data)
	params := map[string]interface{}{
		"event": event,
		"data":  data,
	}
	b, _ := json.Marshal(params)
	req, err := http.NewRequest("POST", url, bytes.NewReader(b))
	if err != nil {
		l.Warning("notify %s: %v", url, err)
		return
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		l.Warning("notify %s %s: %v", event, url, err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		l.Warning("notify %s %s failed: status code %d", event, url, resp.StatusCode)
		return
	}
}
