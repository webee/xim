package service

import (
	"errors"
	"xim/xchat/logic/cache"
	"xim/xchat/logic/service/types"
)

// errors
var (
	ErrInvalidUserStatus = errors.New("invalid user status")
)

// UpdateUserStatus update user's online status.
func UpdateUserStatus(instanceID, sessionID uint64, user string, status string) (err error) {
	var userStatus int
	switch status {
	case types.UserStatusOnline:
		userStatus = cache.StatusOnline
	case types.UserStatusOffline:
		userStatus = cache.StatusOffline
	default:
		return ErrInvalidUserStatus
	}

	cache.UpdateUsers() <- &cache.UserInstanceStatus{
		UserInstance: cache.UserInstance{
			InstanceID: instanceID,
			SessionID:  sessionID,
			User:       user,
		},
		Status: userStatus,
	}
	return
}
