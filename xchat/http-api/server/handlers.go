package server

import (
	"net/http"
	pubtypes "xim/xchat/logic/pub/types"
	"xim/xchat/logic/service"
	"xim/xchat/logic/service/types"

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
	User string `json:"user"`
	Msg  string `json:"msg"`
}

func sendMsg(c echo.Context) error {
	args := &SendMsgArgs{}
	if err := c.Bind(args); err != nil {
		return err
	}

	chatIdentity, err := service.ParseChatIdentity(args.ChatID)
	if err != nil {
		return err
	}
	chatID := chatIdentity.ID
	chatType := chatIdentity.Type

	user := args.User
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
			Kind:     types.MsgKindChat,
		}, &message); err != nil {
			l.Warning("%s error: %s", types.RPCXChatSendMsg, err)
			return c.JSON(http.StatusOK, map[string]interface{}{"ok": false, "error": "send msg failed"})
		}
	case types.MsgKindChatNotify:
		xchatLogic.AsyncCall(types.RPCXChatSendMsg, &types.SendMsgArgs{
			ChatID:   chatID,
			ChatType: chatType,
			User:     user,
			Msg:      msg,
			Kind:     types.MsgKindChatNotify,
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

	user := args.User
	msg := args.Msg
	if len(msg) > 32*1024 {
		return c.JSON(http.StatusOK, map[string]interface{}{"ok": false, "error": "msg excced size limit"})
	}

	var reply types.SendUserNotifyReply
	if err := xchatLogic.Call(types.RPCXChatSendUserNotify, &types.SendUserNotifyArgs{
		User: user,
		Msg:  msg,
	}, &reply); err != nil {
		l.Warning("%s error: %s", types.RPCXChatSendUserNotify, err)
		return c.JSON(http.StatusOK, map[string]interface{}{"ok": false, "error": "send msg failed"})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"ok": reply.Ok})
}
