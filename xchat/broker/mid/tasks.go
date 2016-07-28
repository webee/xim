package mid

import (
	"time"
	"xim/xchat/logic/service/types"

	"gopkg.in/webee/turnpike.v2"
)

// TaskUpdatingOnlineUsers sync online users.
func TaskUpdatingOnlineUsers() {
	for {
		time.Sleep(95 * time.Second)
		// 下线超过30s没有设置client_info（即为初始值"*"）的session.
		sesses := GetSlowSessions()
		for _, id := range sesses {
			realm.KillSession(turnpike.ID(id))
		}

		time.Sleep(5 * time.Second)

		users := GetOnlineSessionUsers()
		xchatLogic.AsyncCall(types.RPCXChatSyncOnlineUsers, &types.SyncOnlineUsersArgs{
			InstanceID: instanceID,
			Users:      users,
		})
	}
}
