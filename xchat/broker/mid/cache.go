package mid

import (
	"strconv"
	"time"
	"xim/xchat/logic/db"
	"xim/xchat/logic/service/types"

	"github.com/patrickmn/go-cache"
)

var (
	chatMembersCache = cache.New(5*time.Minute, 1*time.Minute)
)

func getChatMembers(chatID uint64) []db.Member {
	key := strconv.FormatUint(chatID, 10)
	value, ok := chatMembersCache.Get(key)
	if ok {
		return value.([]db.Member)
	}

	members := []db.Member{}
	if err := xchatLogic.Call(types.RPCXChatFetchChatMembers, chatID, &members); err != nil {
		l.Warning("fetch chat[%d] members error: %s", chatID, err)
		return members
	}
	chatMembersCache.Set(key, members, cache.DefaultExpiration)

	return members
}
