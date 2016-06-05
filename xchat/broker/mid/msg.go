package mid

import (
	"fmt"
	"xim/xchat/logic/db"
	pubtypes "xim/xchat/logic/pub/types"
	"xim/xchat/logic/service/types"
)

func handleMsg(ms <-chan interface{}) {
	for m := range ms {
		switch msg := m.(type) {
		case *pubtypes.Message:
			go push(msg)
		}
	}
}

func push(msg *pubtypes.Message) {
	var members []db.Member
	if err := xchatLogic.Call(types.RPCXChatFetchChatMembers, msg.ChatID, &members); err != nil {
		l.Warning("fetch chat[%d] members error: %s", msg.ChatID, err)
		return
	}
	toPushMsg := NewMessageFromDBMsg(msg)
	for _, member := range members {
		ss := GetUserSessions(member.User)
		for _, x := range ss {
			_ = xchat.Publish(fmt.Sprintf(URIXChatUserMsg, x.ID), []interface{}{toPushMsg}, emptyKwargs)
		}
	}
}
