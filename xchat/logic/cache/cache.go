package cache

import (
	"fmt"
	"os"
	"time"
	"xim/utils/dbutils"
	"xim/utils/netutils"

	"github.com/garyburd/redigo/redis"
)

var (
	keyTimeout    = 120 * time.Second
	redisConnPool *dbutils.RedisConnPool
	toOnline      = make(chan string, 128)
	toOffline     = make(chan string, 128)
)

// InitCache initialize the cache.
func InitCache(redisNetAddr, redisPassword string) (close func()) {
	netAddr, err := netutils.ParseNetAddr(redisNetAddr)
	if err != nil {
		l.Critical("bad redis net addr: %s", redisNetAddr)
		os.Exit(1)
	}
	initRedisConnection(netAddr, redisPassword)

	go running()

	return func() {
		redisConnPool.Close()
	}
}

// InitRedisConnection initialize the redis connection.
func initRedisConnection(netAddr *netutils.NetAddr, password string) {
	redisConnPool = dbutils.NewRedisConnPool(netAddr, password, 64, 128, 30*time.Second)
}

func running() {
	toOnlineUsers := []string{}
	toOfflineUsers := []string{}
	for {
		select {
		case user := <-toOnline:
			toOnlineUsers = append(toOnlineUsers, user)
		case user := <-toOffline:
			toOfflineUsers = append(toOfflineUsers, user)
		case <-time.After(1 * time.Second):
			if len(toOnlineUsers) > 0 {
				userOnline(toOnlineUsers...)
				toOnlineUsers = toOnlineUsers[:0]
			}
			if len(toOfflineUsers) > 0 {
				userOffline(toOfflineUsers...)
				toOfflineUsers = toOfflineUsers[:0]
			}
		}
	}
}

func getUserKey(user string) string {
	return fmt.Sprintf("x.u.%s", user)
}

// userOnline make users online.
func userOnline(users ...string) error {
	redisConn, err := redisConnPool.Get()
	if err != nil {
		return err
	}
	defer redisConnPool.Put(redisConn)

	redisConn.Send("MULTI")
	for _, user := range users {
		userKey := getUserKey(user)
		redisConn.Send("SET", userKey, 1)
		redisConn.Send("EXPIRE", userKey, keyTimeout)
	}
	r, err := redisConn.Do("EXEC")

	l.Debug("user online: %v, %v", users, r)
	return err
}

// userOffline make users offline.
func userOffline(users ...string) error {
	redisConn, err := redisConnPool.Get()
	if err != nil {
		return err
	}
	defer redisConnPool.Put(redisConn)

	redisConn.Send("MULTI")
	for _, user := range users {
		userKey := getUserKey(user)
		redisConn.Send("DEL", userKey)
	}
	r, err := redisConn.Do("EXEC")

	l.Debug("user offline: %v, %v", users, r)
	return err
}

// GetOnlineUsers return online user instances.
func GetOnlineUsers(users ...string) ([]string, error) {
	redisConn, err := redisConnPool.Get()
	if err != nil {
		return nil, err
	}
	defer redisConnPool.Put(redisConn)

	args := redis.Args{}
	for _, user := range users {
		args = args.Add(getUserKey(user))
	}

	vs, err := redis.Values(redisConn.Do("MGET", args...))
	if err != nil {
		return nil, err
	}

	finalUsers := []string{}
	for i, v := range vs {
		if v != nil {
			finalUsers = append(finalUsers, users[i])
		}
	}

	return finalUsers, nil
}
