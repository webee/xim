package service

import (
	"xim/xchat/logic/cache"
	"xim/xchat/logic/service/types"
)

// UpdateUserStatus update user's online status.
func UpdateUserStatus(instanceID, sessionID uint64, user string, status string) {
	var userStatus int
	switch status {
	case types.UserStatusOnline:
		userStatus = cache.StatusOnline
	case types.UserStatusOffline:
		userStatus = cache.StatusOffline
	}
	cache.UpdateUsers() <- &cache.UserInstanceStatus{
		UserInstance: cache.UserInstance{
			InstanceID: instanceID,
			SessionID:  sessionID,
			User:       user,
		},
		Status: userStatus,
	}
}
