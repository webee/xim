package userboard

import (
	"errors"
	"log"
	"sync"
)

// UserBoard records the relations between users and connections.
type UserBoard struct {
	sync.RWMutex
	mapping map[string]map[string]map[string]MsgBroker
}

// NewUserBaord creates a user board.
func NewUserBaord() *UserBoard {
	return &UserBoard{
		mapping: make(map[string]map[string]map[string]MsgBroker),
	}
}

// Register a user.
func (ub *UserBoard) Register(uid *UserIdentity, from string, broker MsgBroker) error {
	var (
		err   error
		ok    bool
		users map[string]map[string]MsgBroker
		froms map[string]MsgBroker
	)
	ub.Lock()
	defer ub.Unlock()

	if users, ok = ub.mapping[uid.Org]; !ok {
		users = make(map[string]map[string]MsgBroker)
		ub.mapping[uid.Org] = users
	}
	if froms, ok = users[uid.User]; !ok {
		froms = make(map[string]MsgBroker)
		users[uid.User] = froms
	}
	if _, ok = froms[from]; !ok {
		froms[from] = broker
	}
	log.Println(uid, from, "registered.")
	// first touch.
	err = ub.Touch(uid, from)
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
		users map[string]map[string]MsgBroker
		froms map[string]MsgBroker
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

// GetUserBroker find the user's broker.
func (ub *UserBoard) GetUserBroker(uid *UserIdentity, from string) (MsgBroker, error) {
	var (
		ok     bool
		users  map[string]map[string]MsgBroker
		froms  map[string]MsgBroker
		broker MsgBroker
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
