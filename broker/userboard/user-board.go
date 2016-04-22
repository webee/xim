package userboard

import (
	"errors"
	"log"
	"sync"
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
func (ub *UserBoard) Register(uid *UserIdentity, instance string, broker UserConn) error {
	var (
		err       error
		ok        bool
		users     map[string]map[string]UserConn
		instances map[string]UserConn
	)
	ub.Lock()
	defer ub.Unlock()

	if users, ok = ub.mapping[uid.Org]; !ok {
		users = make(map[string]map[string]UserConn)
		ub.mapping[uid.Org] = users
	}
	if instances, ok = users[uid.User]; !ok {
		instances = make(map[string]UserConn)
		users[uid.User] = instances
	}
	if _, ok = instances[instance]; !ok {
		instances[instance] = broker
	}
	log.Println(uid, instance, "registered.")
	// first touch.
	err = ub.Touch(uid, instance)
	return err
}

// Touch a user.
func (ub *UserBoard) Touch(uid *UserIdentity, from string) error {
	// reseting redis timeout.
	return nil
}

// Unregister a user.
func (ub *UserBoard) Unregister(uid *UserIdentity, from string) error {
	var (
		ok    bool
		users map[string]map[string]UserConn
		froms map[string]UserConn
	)
	ub.Lock()
	defer ub.Unlock()

	if users, ok = ub.mapping[uid.Org]; ok {
		if froms, ok = users[uid.User]; ok {
			if _, ok = froms[from]; ok {
				delete(froms, from)
			}
			if len(froms) == 0 {
				delete(users, uid.User)
			}
		}
		/*
			if len(users) == 0 {
				delete(ub.mapping, uid.Org)
			}
		*/
	}
	log.Println(uid, from, "unregistered.")
	// delete from redis.
	return nil
}

// GetUserConn find the user's connection.
func (ub *UserBoard) GetUserConn(uid *UserIdentity, from string) (UserConn, error) {
	var (
		ok     bool
		users  map[string]map[string]UserConn
		froms  map[string]UserConn
		broker UserConn
	)
	ub.RLock()
	defer ub.RUnlock()

	if users, ok = ub.mapping[uid.Org]; ok {
		if froms, ok = users[uid.User]; ok {
			if broker, ok = froms[from]; ok {
				return broker, nil
			}
		}
	}
	return nil, errors.New("broker not found")
}
