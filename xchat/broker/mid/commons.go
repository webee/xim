package mid

import (
	"xim/utils/nanorpc"
	"xim/xchat/logic/db"
	pubtypes "xim/xchat/logic/pub/types"
	"xim/xchat/logic/service/types"
)

const (
	// MaxMsgsFetchSize is the max count of msgs been fetched at one time.
	MaxMsgsFetchSize = 3000
)

// FetchChatMsgs fetch chat's messages
func FetchChatMsgs(xchatLogic *nanorpc.Client, user, chatID string, kwargs map[string]interface{}) (msgs []*Message, hasMore, noMore bool, err error) {
	// params
	chatIdentity, err := db.ParseChatIdentity(chatID)
	if err != nil {
		return
	}

	var (
		lid, rid           uint64
		originLimit, limit int
		desc               bool
	)
	if kwargs["lid"] != nil {
		lid = kwargs["lid"].(uint64)
	} else {
		desc = true
	}

	if kwargs["rid"] != nil {
		rid = kwargs["rid"].(uint64)
	}

	if lid > 0 && rid > 0 && lid+1 >= rid {
		return []*Message{}, false, false, nil
	}

	if kwargs["desc"] != nil {
		desc = kwargs["desc"].(bool)
	}

	if kwargs["limit"] != nil {
		limit = kwargs["limit"].(int)
	}
	originLimit = limit
	if lid > 0 && rid > 0 {
		originLimit = int(rid - lid - 1)
	}

	if limit <= 0 {
		limit = 256
	} else if limit > MaxMsgsFetchSize {
		limit = MaxMsgsFetchSize
	}

	var resMsgs []pubtypes.ChatMessage
	method := types.RPCXChatFetchChatMessages
	if user != "" {
		method = types.RPCXChatFetchUserChatMessages
	}
	arguments := &types.FetchUserChatMessagesArgs{
		User:     user,
		ChatID:   chatIdentity.ID,
		ChatType: chatIdentity.Type,
		LID:      lid,
		RID:      rid,
		Limit:    limit,
		Desc:     desc,
	}

	l.Debug("%s: %v", method, arguments)
	if err = xchatLogic.Call(method, arguments, &resMsgs); err != nil {
		return
	}

	msgs = []*Message{}
	for i := range resMsgs {
		msgs = append(msgs, NewMessageFromPubMsg(&resMsgs[i]))
	}
	// 判断是否还有更多数据
	hasMore = false
	if originLimit <= 0 || originLimit > MaxMsgsFetchSize {
		// 没有指定limit或者指定范围超出的情况
		hasMore = len(msgs) >= limit
	}
	noMore = len(msgs) < limit
	return msgs, hasMore, noMore, nil
}
