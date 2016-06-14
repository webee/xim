package mid

import (
	"strconv"
	"time"
	"xim/xchat/logic/db"
	"xim/xchat/logic/service/types"

	"github.com/patrickmn/go-cache"
)

type chatMembers struct {
	updated int64
	members []db.Member
}

var (
	chatMembersCache = cache.New(10*time.Minute, 3*time.Minute)
)

func getChatMembers(chatID uint64, updated int64) []db.Member {
	key := strconv.FormatUint(chatID, 10)
	value, ok := chatMembersCache.Get(key)
	if ok {
		cm := value.(*chatMembers)
		if cm.updated >= updated {
			return cm.members
		}
	}

	members := []db.Member{}
	if err := xchatLogic.Call(types.RPCXChatFetchChatMembers, chatID, &members); err != nil {
		l.Warning("fetch chat[%d] members error: %s", chatID, err)
		return members
	}
	chatMembersCache.Set(key, &chatMembers{
		updated: updated,
		members: members,
	}, cache.DefaultExpiration)

	return members
}
