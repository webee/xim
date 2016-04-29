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
	mapping map[string]map[string]map[string]UserConn
}

// NewUserBaord creates a user board.
func NewUserBaord() *UserBoard {
	return &UserBoard{
		mapping: make(map[string]map[string]map[string]UserConn),
	}
}

// Register a user.
func (ub *UserBoard) Register(user *userds.UserLocation, conn UserConn) error {
	var (
		ok        bool
		users     map[string]map[string]UserConn
		instances map[string]UserConn
	)
	ub.Lock()
	defer ub.Unlock()

	uid := &user.UserIdentity
	instance := user.Instance

	if users, ok = ub.mapping[uid.App]; !ok {
		users = make(map[string]map[string]UserConn)
		ub.mapping[uid.App] = users
	}
	if instances, ok = users[uid.User]; !ok {
		instances = make(map[string]UserConn)
		users[uid.User] = instances
	}
	instances[instance] = conn
	log.Println(uid, instance, "registered.")
	// first touch.
	return userdb.UserOnline(user)
}

// Unregister a user.
func (ub *UserBoard) Unregister(user *userds.UserLocation) error {
	var (
		ok        bool
		users     map[string]map[string]UserConn
		instances map[string]UserConn
	)
	ub.Lock()
	defer ub.Unlock()
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
	log.Println(uid, instance, "unregistered.")
	return userdb.UserOffline(user)
}

// GetUserConn find the user's connection.
func (ub *UserBoard) GetUserConn(user *userds.UserLocation) (UserConn, error) {
	var (
		ok        bool
		users     map[string]map[string]UserConn
		instances map[string]UserConn
		broker    UserConn
	)
	ub.RLock()
	defer ub.RUnlock()
	uid := &user.UserIdentity
	instance := user.Instance

	if users, ok = ub.mapping[uid.App]; ok {
		if instances, ok = users[uid.User]; ok {
			if broker, ok = instances[instance]; ok {
				return broker, nil
			}
		}
	}
	return nil, errors.New("broker not found")
}
