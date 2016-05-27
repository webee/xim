package userboard

import (
	"errors"
	"log"
	"sync"
	"xim/broker/userdb"
	"xim/broker/userds"
)

// UserBoard records the relations between users and connections.
type UserBoard struct {
	sync.RWMutex
	mapping map[string]map[string]map[uint32]UserMsgBox
}

// NewUserBoard creates a user board.
func NewUserBoard() *UserBoard {
	return &UserBoard{
		mapping: make(map[string]map[string]map[uint32]UserMsgBox),
	}
}

// Register a user.
func (ub *UserBoard) Register(user *userds.UserLocation, conn UserMsgBox) error {
	var (
		ok        bool
		users     map[string]map[uint32]UserMsgBox
		instances map[uint32]UserMsgBox
	)
	ub.Lock()

	uid := &user.UserIdentity
	instance := user.Instance

	if users, ok = ub.mapping[uid.App]; !ok {
		users = make(map[string]map[uint32]UserMsgBox)
		ub.mapping[uid.App] = users
	}
	if instances, ok = users[uid.User]; !ok {
		instances = make(map[uint32]UserMsgBox)
		users[uid.User] = instances
	}
	instances[instance] = conn
	log.Println(uid, instance, "registered.")
	// first touch.
	ub.Unlock()

	return userdb.UserOnline(user)
}

// Unregister a user.
func (ub *UserBoard) Unregister(user *userds.UserLocation) error {
	var (
		ok        bool
		users     map[string]map[uint32]UserMsgBox
		instances map[uint32]UserMsgBox
	)
	ub.Lock()
	uid := &user.UserIdentity
	instance := user.Instance

	if users, ok = ub.mapping[uid.App]; ok {
		if instances, ok = users[uid.User]; ok {
			if _, ok = instances[instance]; ok {
				delete(instances, instance)
			}
			if len(instances) == 0 {
				delete(users, uid.User)
			}
		}
		/*
			if len(users) == 0 {
				delete(ub.mapping, uid.App)
			}
		*/
	}
	ub.Unlock()
	log.Println(uid, instance, "unregistered.")
	return userdb.UserOffline(user)
}

// GetUserMsgBox find the user's msgbox.
func (ub *UserBoard) GetUserMsgBox(user *userds.UserLocation) (UserMsgBox, error) {
	var (
		ok        bool
		users     map[string]map[uint32]UserMsgBox
		instances map[uint32]UserMsgBox
		msgBox    UserMsgBox
	)
	ub.RLock()
	defer ub.RUnlock()
	uid := &user.UserIdentity
	instance := user.Instance

	if users, ok = ub.mapping[uid.App]; ok {
		if instances, ok = users[uid.User]; ok {
			if msgBox, ok = instances[instance]; ok {
				return msgBox, nil
			}
		}
	}
	return nil, errors.New("broker not found")
}
