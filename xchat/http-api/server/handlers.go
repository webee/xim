package server

import (
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
	ChatID    string `json:"chat_id"`
	User      string `json:"user"`
	Domain    string `json:"domain"`
	Msg       string `json:"msg"`
	Kind      string `json:"kind"`
	PermCheck bool   `json:"perm_check"`
}

// SendUserNotifyArgs is arguments of sendUserNotify.
type SendUserNotifyArgs struct {
	User      string `json:"user"`
	ToUser    string `json:"to_user"`
	Domain    string `json:"domain"`
	Msg       string `json:"msg"`
	PermCheck bool   `json:"perm_check"`
}

func getNs(c echo.Context) string {
	return GetContextString(NsContextKey, c)
}

func getNsUser(c echo.Context, u string) string {
	ns := getNs(c)
	return nsutils.EncodeNSUser(ns, u)
}

func sendMsg(c echo.Context) error {
	args := &SendMsgArgs{}
	if err := c.Bind(args); err != nil {
		return err
	}

	chatIdentity, err := db.ParseChatIdentity(args.ChatID)
	if err != nil {
		return err
	}
	chatID := chatIdentity.ID
	chatType := chatIdentity.Type
	domain := args.Domain

	user := getNsUser(c, args.User)
	msg := args.Msg
	if len(msg) > 64*1024 {
		return c.JSON(http.StatusOK, map[string]interface{}{"ok": false, "error": "msg excced size limit"})
	}
	ignorePermCheck := !args.PermCheck
	sendMsgArgs := &types.SendMsgArgs{
		ChatID:   chatID,
		ChatType: chatType,
		Domain:   domain,
		User:     user,
		Msg:      msg,
		Options: &types.SendMsgOptions{
			IgnorePermCheck: ignorePermCheck,
		},
	}

	switch args.Kind {
	case "", types.MsgKindChat:
		var message pubtypes.ChatMessage
		if err := xchatLogic.Call(types.RPCXChatSendMsg, sendMsgArgs, &message); err != nil {
			l.Warning("%s error: %s", types.RPCXChatSendMsg, err)
			return c.JSON(http.StatusOK, map[string]interface{}{"ok": false, "error": "send msg failed"})
		}
	case types.MsgKindChatNotify:
		xchatLogic.AsyncCall(types.RPCXChatSendNotify, sendMsgArgs)
	default:
		return c.JSON(http.StatusOK, map[string]interface{}{"ok": false, "error": "invalid msg kind"})
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
