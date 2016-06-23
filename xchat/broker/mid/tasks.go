package mid

import (
	"time"
	"xim/xchat/logic/service/types"
)

// TaskUpdatingOnlineUsers sync online users.
func TaskUpdatingOnlineUsers() {
	for {
		time.Sleep(100 * time.Second)
		users := GetOnlineSessionUsers()
		xchatLogic.AsyncCall(types.RPCXChatSyncOnlineUsers, &types.SyncOnlineUsersArgs{
			InstanceID: instanceID,
			Users:      users,
		})
	}
}
