package userdb

import (
	"fmt"
	"log"
	"time"
	"xim/utils/dbutils"
	"xim/utils/netutils"

	"xim/broker/userds"

	"github.com/garyburd/redigo/redis"
)

var (
	redisConnPool *dbutils.RedisConnPool
	//toOnlineUsers  chan *userds.UserLocation
	//toOfflineUsers chan *userds.UserLocation
)

// InitUserDB initialize the user db.
func InitUserDB(c *Config) {
	initConfig(c)

	netAddr, err := netutils.ParseNetAddr(config.RedisNetAddr)
	if err != nil {
		log.Fatalln("bad redis net addr:", config.RedisNetAddr)
	}
	initRedisConnection(netAddr, c.RedisPassword)

	/*
		toOnlineUsers = make(chan *userds.UserLocation, 1000)
		toOfflineUsers = make(chan *userds.UserLocation, 1000)
		// working.
		go onlineWorker()
		go offlineWorker()
	*/
}

/*
func onlineWorker() {
	users := []*userds.UserLocation{}
	for {
		select {
		case user := <-toOnlineUsers:
			users = append(users, user)
		case <-time.After(1 * time.Second):
			if len(users) > 0 {
				_ = userOnline(users...)
			}
			users = []*userds.UserLocation{}
		}
	}
}

func offlineWorker() {
	users := []*userds.UserLocation{}
	for {
		select {
		case user := <-toOfflineUsers:
			users = append(users, user)
		case <-time.After(1 * time.Second):
			if len(users) > 0 {
				_ = userOffline(users...)
			}
			users = []*userds.UserLocation{}
		}
	}
}
*/

// InitRedisConnection initialize the redis connection.
func initRedisConnection(netAddr *netutils.NetAddr, password string) {
	redisConnPool = dbutils.NewRedisConnPool(netAddr, password, 32, 64, 30*time.Second)
}

func getUserKey(uid *userds.UserIdentity) string {
	return fmt.Sprintf("x.u.%s", uid.String())
}

func getUserLocationKey(user *userds.UserLocation) string {
	return fmt.Sprintf("x.u.l.%s", user.String())
}

/*
// UserOnline make a user online.
func UserOnline(user *userds.UserLocation) {
	toOnlineUsers <- user
}

// UserOffline make a user offline.
func UserOffline(user *userds.UserLocation) {
	toOfflineUsers <- user
}
*/

// UserOnline make a user online.
func UserOnline(users ...*userds.UserLocation) error {
	redisConn, err := redisConnPool.Get()
	if err != nil {
		return err
	}
	defer redisConnPool.Put(redisConn)

	for _, user := range users {
		userKey := getUserKey(&user.UserIdentity)
		userLocationKey := getUserLocationKey(user)
		value := user.String()
		redisConn.Send("MULTI")
		redisConn.Send("SADD", userKey, value)
		redisConn.Send("EXPIRE", userKey, config.UserTimeout)
		redisConn.Send("SET", userLocationKey, 1)
		redisConn.Send("EXPIRE", userLocationKey, config.UserTimeout)
		r, _ := redisConn.Do("EXEC")

		log.Println("user online:", user, r)
	}
	return err
}

// UserOffline make a user offline.
func UserOffline(users ...*userds.UserLocation) error {
	redisConn, err := redisConnPool.Get()
	if err != nil {
		return err
	}
	defer redisConnPool.Put(redisConn)

	for _, user := range users {
		userKey := getUserKey(&user.UserIdentity)
		userLocationKey := getUserLocationKey(user)
		value := user.String()
		redisConn.Send("MULTI")
		redisConn.Send("SREM", userKey, value)
		redisConn.Send("DEL", userLocationKey)
		r, _ := redisConn.Do("EXEC")

		log.Println("user offline:", user, r)
	}
	return err
}

// GetOnlineUsers return online user instances.
func GetOnlineUsers(uids ...*userds.UserIdentity) ([]*userds.UserLocation, error) {
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

	users := []*userds.UserLocation{}
	args = redis.Args{}
	for _, r := range rs {
		user := userds.ParseUserLocation(r)
		users = append(users, user)
		args = args.Add(getUserLocationKey(user))
	}

	vs, err := redis.Values(redisConn.Do("MGET", args...))
	if err != nil {
		return nil, err
	}

	log.Println(users)
	finalUsers := []*userds.UserLocation{}
	for i, v := range vs {
		if v != nil {
			finalUsers = append(finalUsers, users[i])
		}
	}

	return finalUsers, nil
}
