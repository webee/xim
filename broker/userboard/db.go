package userboard

import (
	"fmt"
	"log"
	"time"
	"xim/utils/dbutils"
	"xim/utils/netutils"

	"github.com/garyburd/redigo/redis"
)

var (
	redisConnPool *dbutils.RedisConnPool
)

// InitRedisConnection initialize the redis connection.
func initRedisConnection(netAddr *netutils.NetAddr) {
	redisConnPool = dbutils.NewRedisConnPool(netAddr, 4, 8, time.Second)
}

func getUserKey(uid *UserIdentity) string {
	return fmt.Sprintf("x.u.%s", uid.String())
}

func getUserLocationKey(user *UserLocation) string {
	return fmt.Sprintf("x.u.l.%s", user.String())
}

// UserOnline make a user online.
func UserOnline(user *UserLocation) error {
	redisConn, err := redisConnPool.Get()
	if err != nil {
		return err
	}
	defer redisConnPool.Put(redisConn)

	userKey := getUserKey(&user.UserIdentity)
	userLocationKey := getUserLocationKey(user)
	value := user.String()
	redisConn.Send("MULTI")
	redisConn.Send("SADD", userKey, value)
	redisConn.Send("EXPIRE", userKey, config.UserTimeout)
	redisConn.Send("SET", userLocationKey, 1)
	redisConn.Send("EXPIRE", userLocationKey, config.UserTimeout)
	r, err := redisConn.Do("EXEC")

	log.Println("user online:", user, r)
	return err
}

// UserOffline make a user offline.
func UserOffline(user *UserLocation) error {
	redisConn, err := redisConnPool.Get()
	if err != nil {
		return err
	}
	defer redisConnPool.Put(redisConn)

	userKey := getUserKey(&user.UserIdentity)
	userLocationKey := getUserLocationKey(user)
	value := user.String()
	redisConn.Send("MULTI")
	redisConn.Send("SREM", userKey, value)
	redisConn.Send("DEL", userLocationKey)
	r, err := redisConn.Do("EXEC")

	log.Println("user offline:", user, r)
	return err
}

// GetOnlineUsers returns the user locations which are onlines.
func GetOnlineUsers(uids ...*UserIdentity) ([]*UserLocation, error) {
	redisConn, err := redisConnPool.Get()
	if err != nil {
		return nil, err
	}
	defer redisConnPool.Put(redisConn)

	args := redis.Args{}
	for _, uid := range uids {
		args = args.Add(getUserKey(uid))
	}

	rs, err := redis.Strings(redisConn.Do("SUNION", args...))
	if err != nil {
		return nil, err
	}

	users := []*UserLocation{}
	args = redis.Args{}
	for _, r := range rs {
		user := ParseUserLocation(r)
		users = append(users, user)
		args = args.Add(getUserLocationKey(user))
	}

	vs, err := redis.Values(redisConn.Do("MGET", args...))
	if err != nil {
		return nil, err
	}

	log.Println(users)
	finalUsers := []*UserLocation{}
	for i, v := range vs {
		if v != nil {
			finalUsers = append(finalUsers, users[i])
		}
	}

	return finalUsers, nil
}
