package server

import (
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

	user := getNsUser(c, args.User)
	msg := args.Msg
	if len(msg) > 64*1024 {
		return c.JSON(http.StatusOK, map[string]interface{}{"ok": false, "error": "msg excced size limit"})
	}
	ignorePermCheck := !args.PermCheck
	sendMsgArgs := &types.SendMsgArgs{
		ChatID:   chatID,
		ChatType: chatType,
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
	user := nsutils.EncodeNSUser(ns, args.User)
	toUser := nsutils.EncodeNSUser(ns, args.ToUser)
	if ns == "notify" {
		// notify 可以发送给任何人，其它ns则只能给自己的ns用户发送, 当然空ns也可以给任何人发送
		toUser = args.ToUser
	}

	msg := args.Msg
	if len(msg) > 32*1024 {
		return c.JSON(http.StatusOK, map[string]interface{}{"ok": false, "error": "msg excced size limit"})
	}

	var ts int64
	ignorePermCheck := !args.PermCheck
	if err := xchatLogic.Call(types.RPCXChatSendUserNotify, &types.SendUserMsgArgs{
		ToUser: toUser,
		User:   user,
		Domain: args.Domain,
		Msg:    msg,
		Options: &types.SendMsgOptions{
			IgnorePermCheck: ignorePermCheck,
		},
	}, &ts); err != nil {
		l.Warning("%s error: %s", types.RPCXChatSendUserNotify, err)
		return c.JSON(http.StatusOK, map[string]interface{}{"ok": false, "error": "send msg failed"})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"ok": true, "ts": ts})
}
