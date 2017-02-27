package server

import (
	"errors"
	"fmt"
	"net/http"
	"xim/xchat/logic/db"
	pubtypes "xim/xchat/logic/pub/types"
	"xim/xchat/logic/service/types"

	"xim/utils/nsutils"

	"github.com/labstack/echo"
)

// SendMsgArgs is arguments of sendMsg.
type SendMsgArgs struct {
	ChatID           string   `json:"chat_id"`
	User             string   `json:"user"`
	Domain           string   `json:"domain"`
	Msg              string   `json:"msg"`
	Kind             string   `json:"kind"`
	ForceNotifyUsers []string `json:"force_notify_users"`
	PermCheck        bool     `json:"perm_check"`
}

// SendUniqueChatMsgArgs is arguments of sendUniqueChatMsg.
type SendUniqueChatMsgArgs struct {
	ChatType string   `json:"chat_type"`
	User     string   `json:"user"`
	ToUsers  []string `json:"to_users"`
	Domain   string   `json:"domain"`
	Msg      string   `json:"msg"`
	Kind     string   `json:"kind"`
}

// SendUserNotifyArgs is arguments of sendUserNotify.
type SendUserNotifyArgs struct {
	User      string `json:"user"`
	ToUser    string `json:"to_user"`
	Domain    string `json:"domain"`
	Msg       string `json:"msg"`
	PermCheck bool   `json:"perm_check"`
}

var (
	emptyStruct = struct{}{}
)

func getNs(c echo.Context) string {
	return GetContextString(NsContextKey, c)
}

func doSendMsg(kind, domain, xchatID, msg, user string, forceNotifyUsers map[string]struct{}, permCheck bool) (id uint64, ts int64, err error) {
	if len(msg) > 64*1024 {
		err = errors.New("msg excced size limit")
		return
	}

	chatIdentity, err := db.ParseChatIdentity(xchatID)
	if err != nil {
		return
	}
	chatID := chatIdentity.ID
	chatType := chatIdentity.Type
	ignorePermCheck := !permCheck
	sendMsgArgs := &types.SendMsgArgs{
		ChatID:           chatID,
		ChatType:         chatType,
		Domain:           domain,
		User:             user,
		Msg:              msg,
		ForceNotifyUsers: forceNotifyUsers,
		Options: &types.SendMsgOptions{
			IgnorePermCheck: ignorePermCheck,
		},
	}

	switch kind {
	case "", types.MsgKindChat:
		var message pubtypes.ChatMessage
		if errx := xchatLogic.Call(types.RPCXChatSendMsg, sendMsgArgs, &message); errx != nil {
			l.Warning("%s error: %s", types.RPCXChatSendMsg, errx)
			err = errors.New("send msg failed")
			return
		}
		id = message.ID
		ts = message.Ts
	case types.MsgKindChatNotify:
		xchatLogic.AsyncCall(types.RPCXChatSendNotify, sendMsgArgs)
	default:
		err = errors.New("invalid msg kind")
		return
	}
	return
}

func sendMsg(c echo.Context) error {
	args := &SendMsgArgs{}
	if err := c.Bind(args); err != nil {
		return err
	}

	ns := getNs(c)
	user := nsutils.EncodeNSUser(ns, args.User)
	forceNotifyUsers := make(map[string]struct{})
	for _, u := range args.ForceNotifyUsers {
		forceNotifyUsers[nsutils.EncodeNSUser(ns, u)] = emptyStruct
	}

	id, ts, err := doSendMsg(args.Kind, args.Domain, args.ChatID, args.Msg, user, forceNotifyUsers, args.PermCheck)
	if err != nil {
		return c.JSON(http.StatusOK, map[string]interface{}{"ok": false, "error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"ok": true, "id": id, "ts": ts})
}

func sendUniqueChatMsg(c echo.Context) error {
	args := &SendUniqueChatMsgArgs{}
	if err := c.Bind(args); err != nil {
		return err
	}
	ns := getNs(c)
	user := nsutils.EncodeNSUser(ns, args.User)
	// 判断chatType
	users := []string{user}
	chatType := ""
	if args.ToUsers == nil {
		chatType = "self"
	} else {
		switch len(args.ToUsers) {
		case 1:
			if args.ToUsers[0] == "cs:*" {
				chatType = "cs"
			} else {
				chatType = "user"
				users = append(users, nsutils.EncodeNSUser(ns, args.ToUsers[0]))
			}
		case 0:
			chatType = "self"
		default:
			return c.JSON(http.StatusOK, map[string]interface{}{"ok": false, "error": "bad param: to_users"})
		}
	}
	if args.ChatType != "" && args.ChatType != chatType {
		return c.JSON(http.StatusOK, map[string]interface{}{"ok": false, "error": "bad param: chat_type"})
	}

	xchatID, err := xchatHTTPClient.NewChat(chatType, users, "", "user", "")
	if err != nil {
		return err
	}

	id, ts, err := doSendMsg(args.Kind, args.Domain, xchatID, args.Msg, user, nil, false)
	if err != nil {
		return c.JSON(http.StatusOK, map[string]interface{}{"ok": false, "error": err.Error()})
	}
	if id > 0 {
		return c.JSON(http.StatusOK, map[string]interface{}{"ok": true, "id": id, "ts": ts})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"ok": true})
}

func sendUserNotify(c echo.Context) error {
	args := &SendUserNotifyArgs{}
	if err := c.Bind(args); err != nil {
		return err
	}

	ns := getNs(c)
	user := args.User
	toUser := args.ToUser
	if toUser == "" && user != "" {
		// TODO: remove
		// 兼容老接口
		user = nsutils.EncodeNSUser(ns, "")
		toUser = nsutils.EncodeNSUser(ns, args.User)
		if ns == "notify" {
			// notify 可以发送给任何人，其它ns则只能给自己的ns用户发送, 当然空ns也可以给任何人发送
			toUser = args.User
		}
	} else {
		user = nsutils.EncodeNSUser(ns, user)
		toUser = nsutils.EncodeNSUser(ns, toUser)
		if ns == "notify" {
			// notify 可以发送给任何人，其它ns则只能给自己的ns用户发送, 当然空ns也可以给任何人发送
			toUser = args.ToUser
		}
	}

	msg := args.Msg
	if len(msg) > 32*1024 {
		return c.JSON(http.StatusOK, map[string]interface{}{"ok": false, "error": "msg excced size limit"})
	}

	var ts int64
	ignorePermCheck := user == "" || !args.PermCheck
	if err := xchatLogic.Call(types.RPCXChatSendUserNotify, &types.SendUserMsgArgs{
		ToUser: toUser,
		User:   user,
		Domain: args.Domain,
		Msg:    msg,
		Options: &types.SendMsgOptions{
			IgnorePermCheck: ignorePermCheck,
		},
	}, &ts); err != nil {
		errMsg := fmt.Sprintf("send msg failed: %s", err)
		l.Warning("%s, %s", types.RPCXChatSendUserNotify, errMsg)
		return c.JSON(http.StatusOK, map[string]interface{}{"ok": false, "error": errMsg})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"ok": true, "ts": ts})
}
