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
	ChatID string `json:"chat_id"`
	User   string `json:"user"`
	Msg    string `json:"msg"`
	Kind   string `json:"kind"`
}

// SendUserNotifyArgs is arguments of sendUserNotify.
type SendUserNotifyArgs struct {
	User   string `json:"user"`
	Domain string `json:"domain"`
	Msg    string `json:"msg"`
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

	switch args.Kind {
	case "", types.MsgKindChat:
		var message pubtypes.ChatMessage
		if err := xchatLogic.Call(types.RPCXChatSendMsg, &types.SendMsgArgs{
			ChatID:   chatID,
			ChatType: chatType,
			User:     user,
			Msg:      msg,
		}, &message); err != nil {
			l.Warning("%s error: %s", types.RPCXChatSendMsg, err)
			return c.JSON(http.StatusOK, map[string]interface{}{"ok": false, "error": "send msg failed"})
		}
	case types.MsgKindChatNotify:
		xchatLogic.AsyncCall(types.RPCXChatSendNotify, &types.SendMsgArgs{
			ChatID:   chatID,
			ChatType: chatType,
			User:     user,
			Msg:      msg,
		})
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
	if ns == "notify" {
		// notify 可以发送给任何人，其它ns则只能给自己的ns用户发送
		user = args.User
	}

	msg := args.Msg
	if len(msg) > 32*1024 {
		return c.JSON(http.StatusOK, map[string]interface{}{"ok": false, "error": "msg excced size limit"})
	}

	var ts int64
	if err := xchatLogic.Call(types.RPCXChatSendUserNotify, &types.SendUserMsgArgs{
		User:   user,
		Domain: args.Domain,
		Msg:    msg,
	}, &ts); err != nil {
		l.Warning("%s error: %s", types.RPCXChatSendUserNotify, err)
		return c.JSON(http.StatusOK, map[string]interface{}{"ok": false, "error": "send msg failed"})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"ok": true, "ts": ts})
}
