package userboard

import (
	"errors"
	"fmt"
	"log"
	"sync"
)

// UserLocation represents a user connection location.
type UserLocation struct {
	UserIdentity
	Broker   string
	Instance string
}

func (u UserLocation) String() string {
	return fmt.Sprintf("%s:%s>%s#%s", u.Org, u.User, u.Broker, u.Instance)
}

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
func (ub *UserBoard) Register(user *UserLocation, broker UserConn) error {
	var (
		err       error
		ok        bool
		users     map[string]map[string]UserConn
		instances map[string]UserConn
	)
	ub.Lock()
	defer ub.Unlock()

	uid := &user.UserIdentity
	instance := user.Instance

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
	err = ub.Touch(user)
	return err
}

// Touch a user.
func (ub *UserBoard) Touch(user *UserLocation) error {
	// TODO
	// reseting redis timeout.
	log.Println("touch:", user)
	return nil
}

// Unregister a user.
func (ub *UserBoard) Unregister(user *UserLocation) error {
	var (
		ok        bool
		users     map[string]map[string]UserConn
		instances map[string]UserConn
	)
	ub.Lock()
	defer ub.Unlock()
	uid := &user.UserIdentity
	instance := user.Instance

	if users, ok = ub.mapping[uid.Org]; ok {
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
				delete(ub.mapping, uid.Org)
			}
		*/
	}
	log.Println(uid, instance, "unregistered.")
	// delete from redis.
	return nil
}

// GetUserConn find the user's connection.
func (ub *UserBoard) GetUserConn(user *UserLocation) (UserConn, error) {
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

	if users, ok = ub.mapping[uid.Org]; ok {
		if instances, ok = users[uid.User]; ok {
			if broker, ok = instances[instance]; ok {
				return broker, nil
			}
		}
	}
	return nil, errors.New("broker not found")
}
