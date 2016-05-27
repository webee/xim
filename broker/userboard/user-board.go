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
	mapping map[uint32]UserMsgBox
	acts    chan func()
}

// NewUserBoard creates a user board.
func NewUserBoard() *UserBoard {
	return &UserBoard{
		mapping: make(map[uint32]UserMsgBox),
		acts:    make(chan func()),
	}
}

// Register a user.
func (ub *UserBoard) Register(user *userds.UserLocation, conn UserMsgBox) error {
	ub.Lock()
	instance := user.Instance
	if _, ok := ub.mapping[instance]; !ok {
		ub.mapping[instance] = conn
	}
	defer ub.Unlock()
	log.Println(user, "registered.")

	return userdb.UserOnline(user)
}

// Unregister a user.
func (ub *UserBoard) Unregister(user *userds.UserLocation) error {
	ub.Lock()
	instance := user.Instance
	if _, ok := ub.mapping[instance]; ok {
		delete(ub.mapping, instance)
	}
	defer ub.Unlock()

	log.Println(user, "unregistered.")
	return userdb.UserOffline(user)
}

// GetUserMsgBox get the user's msg box.
func (ub *UserBoard) GetUserMsgBox(user *userds.UserLocation) (UserMsgBox, error) {
	ub.RLock()
	defer ub.RUnlock()
	instance := user.Instance
	if conn, ok := ub.mapping[instance]; ok {
		return conn, nil
	}
	return nil, errors.New("broker not found")
}
