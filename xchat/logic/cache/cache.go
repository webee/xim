package cache

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"xim/utils/dbutils"
	"xim/utils/netutils"

	"github.com/garyburd/redigo/redis"
)

// errors
var (
	ErrInvalidUserInstanceKey = errors.New("invalid user instance key")
)

// constants.
const (
	StatusOnline  = 0
	StatusOffline = 1
)

// UserInstance is a user instance.
type UserInstance struct {
	InstanceID uint64
	SessionID  uint64
	User       string
}

// UserInstanceStatus is user instance's status.
type UserInstanceStatus struct {
	UserInstance
	Status int
}

func (ui *UserInstance) String() string {
	return fmt.Sprintf("%s.%s.%s", strconv.FormatUint(ui.InstanceID, 36), strconv.FormatUint(ui.SessionID, 36), ui.User)
}

func genUserKey(user string) string {
	return fmt.Sprintf("x.u.%s", user)
}

func genUserInstanceKey(user string, instanceID, sessionID uint64) string {
	return fmt.Sprintf("x.i.%s.%s.%s", strconv.FormatUint(instanceID, 36), strconv.FormatUint(sessionID, 36), user)
}

// ParseUserInstanceFromKey parse user instance from user instance key.
func ParseUserInstanceFromKey(s string) (*UserInstance, error) {
	parts := strings.SplitN(s, ".", 5)
	if len(parts) < 4 {
		return nil, ErrInvalidUserInstanceKey
	}
	instanceID, err := strconv.ParseUint(parts[2], 36, 64)
	if err != nil {
		return nil, err
	}
	sessionID, err := strconv.ParseUint(parts[3], 36, 64)
	if err != nil {
		return nil, err
	}

	user := parts[4]
	return &UserInstance{
		InstanceID: instanceID,
		SessionID:  sessionID,
		User:       user,
	}, nil
}

var (
	keyTimeout    = 120
	redisConnPool *dbutils.RedisConnPool
	toUpdate      = make(chan *UserInstanceStatus, 512)
)

// InitCache initialize the cache.
func InitCache(redisNetAddr, redisPassword string, db int, poolSize int) (close func()) {
	netAddr, err := netutils.ParseNetAddr(redisNetAddr)
	if err != nil {
		l.Critical("bad redis net addr: %s", redisNetAddr)
		os.Exit(1)
	}
	redisConnPool = dbutils.NewRedisConnPool(netAddr, redisPassword, db, poolSize, 2*poolSize, 30*time.Second)

	go running()

	return func() {
		redisConnPool.Close()
	}
}

func running() {
	toOnlineUsers := map[string]*UserInstance{}
	toOfflineUsers := map[string]*UserInstance{}
	for {
		select {
		case userStatus := <-toUpdate:
			switch userStatus.Status {
			case StatusOnline:
				userInstanceKey := genUserInstanceKey(userStatus.User, userStatus.InstanceID, userStatus.SessionID)
				toOnlineUsers[userInstanceKey] = &userStatus.UserInstance
				delete(toOfflineUsers, userInstanceKey)
			case StatusOffline:
				userInstanceKey := genUserInstanceKey(userStatus.User, userStatus.InstanceID, userStatus.SessionID)
				toOfflineUsers[userInstanceKey] = &userStatus.UserInstance
				delete(toOnlineUsers, userInstanceKey)
			}
			if len(toOnlineUsers) > 256 {
				userOnline(toOnlineUsers)
				toOnlineUsers = map[string]*UserInstance{}
			}
			if len(toOfflineUsers) > 256 {
				userOffline(toOfflineUsers)
				toOfflineUsers = map[string]*UserInstance{}
			}
		case <-time.After(2 * time.Second):
			if len(toOnlineUsers) > 0 {
				userOnline(toOnlineUsers)
				toOnlineUsers = map[string]*UserInstance{}
			}
			if len(toOfflineUsers) > 0 {
				userOffline(toOfflineUsers)
				toOfflineUsers = map[string]*UserInstance{}
			}
		}
	}
}

// UpdateUsers is the to update users channel.
func UpdateUsers() chan<- *UserInstanceStatus {
	return toUpdate
}

// userOnline make users online.
func userOnline(users map[string]*UserInstance) error {
	redisConn, err := redisConnPool.Get()
	if err != nil {
		l.Warning("get redis conn error: %s", err.Error())
		return err
	}
	defer redisConnPool.Put(redisConn)

	redisConn.Send("MULTI")
	for userInstanceKey, user := range users {
		userKey := genUserKey(user.User)
		redisConn.Send("SADD", userKey, userInstanceKey)
		redisConn.Send("EXPIRE", userKey, keyTimeout)
		redisConn.Send("SET", userInstanceKey, 1)
		redisConn.Send("EXPIRE", userInstanceKey, keyTimeout)
	}
	r, err := redisConn.Do("EXEC")

	l.Debug("user online: %v, %v", users, r)
	return err
}

// userOffline make users offline.
func userOffline(users map[string]*UserInstance) error {
	redisConn, err := redisConnPool.Get()
	if err != nil {
		l.Warning("get redis conn error: %s", err.Error())
		return err
	}
	defer redisConnPool.Put(redisConn)

	redisConn.Send("MULTI")
	for userInstanceKey, user := range users {
		userKey := genUserKey(user.User)
		redisConn.Send("SREM", userKey, userInstanceKey)
		redisConn.Send("DEL", userInstanceKey)
	}
	r, err := redisConn.Do("EXEC")

	l.Debug("user offline: %v, %v", users, r)
	return err
}

func doGetOnlineUsers(users []string) ([]interface{}, []*UserInstance, error) {
	if len(users) == 0 {
		return nil, nil, nil
	}

	redisConn, err := redisConnPool.Get()
	if err != nil {
		return nil, nil, err
	}
	defer redisConnPool.Put(redisConn)

	args := redis.Args{}
	for _, user := range users {
		args = args.Add(genUserKey(user))
	}

	rs, err := redis.Strings(redisConn.Do("SUNION", args...))
	if err != nil {
		return nil, nil, err
	}

	userInstances := []*UserInstance{}
	args = redis.Args{}
	for _, r := range rs {
		userInstance, err2 := ParseUserInstanceFromKey(r)
		if err2 != nil {
			l.Warning("parse user instance error: %s", err2.Error())
			continue
		}

		userInstances = append(userInstances, userInstance)
		args = args.Add(r)
	}

	if len(args) == 0 {
		return nil, nil, nil
	}
	vs, err := redis.Values(redisConn.Do("MGET", args...))
	if err != nil {
		return nil, nil, err
	}

	return vs, userInstances, nil
}

// GetOnlineUsers return online user instances.
func GetOnlineUsers(users ...string) ([]*UserInstance, error) {
	vs, userInstances, err := doGetOnlineUsers(users)
	if err != nil {
		return nil, err
	}

	finalUsers := []*UserInstance{}
	for i, v := range vs {
		if v != nil {
			finalUsers = append(finalUsers, userInstances[i])
		}
	}

	return finalUsers, nil
}

// GetOfflineUsers return offline users.
func GetOfflineUsers(users ...string) ([]string, error) {
	vs, userInstances, err := doGetOnlineUsers(users)
	if err != nil {
		return nil, err
	}

	finalUsers := map[string]bool{}
	for i, v := range vs {
		if v != nil {
			finalUsers[userInstances[i].User] = true
		}
	}

	offlineUsers := []string{}
	for _, user := range users {
		if !finalUsers[user] {
			offlineUsers = append(offlineUsers, user)
		}
	}

	return offlineUsers, nil
}
