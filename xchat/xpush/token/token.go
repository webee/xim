package token

import (
	"encoding/json"
	"log"
	"time"
	"xim/xchat/xpush/kafka"
	"errors"
	"xim/utils/dbutils"
	"xim/utils/netutils"
)

const (
	USER_DEV_INFO_KEY = "user:dev:info"
)
var (
	redisConnPool *dbutils.RedisConnPool
)

func InitRedisPool(addr, password string) {
	netAddr := &netutils.NetAddr{Network:"tcp", LAddr:addr}
	redisConnPool = dbutils.NewRedisConnPool(netAddr, password, 32, 64, 30 * time.Second)
}

func GetUserDeviceInfo(addr, user string) (*kafka.UserDeviceInfo, error) {
	log.Println("GetUserDeviceInfo", addr, user)
	ret := &kafka.UserDeviceInfo{}

	conn, err := redisConnPool.Get()
	if err != nil {
		log.Println("redisConnPool.Get failed.", err)
		return ret, err
	}
	defer redisConnPool.Put(conn)

	reply, err := conn.Do("hget", USER_DEV_INFO_KEY, user)
	if err != nil {
		log.Println("redis.Send failed.", err)
		return ret, err
	}

	if reply == nil {
		log.Println("user device info not found.", user)
		return  ret, errors.New("user device info not found.")
	}

	var data []byte
	switch v := reply.(type) {
	case string:
		data = []byte(v)
	case []byte:
		data = v
	}

	log.Println("GetUserDeviceInfo", string(data))

	err = json.Unmarshal(data, &ret)
	if err != nil {
		log.Println("json.Unmarshal failed.", err)
		return ret, err
	}

	return ret, nil
}

func SetUserDeviceInfo(addr, user string, udi *kafka.UserDeviceInfo) error {
	log.Println("SetUserDeviceInfo", user, udi)
	conn, err := redisConnPool.Get()
	if err != nil {
		log.Println("redisConnPool.Get failed.", err)
		return err
	}
	defer redisConnPool.Put(conn)

	json, err := json.Marshal(udi)
	if err != nil {
		log.Println("json.Marshal failed.", err)
		return err
	}
	reply, err := conn.Do("hset", USER_DEV_INFO_KEY, user, string(json))
	if err != nil {
		log.Println("redis.Send failed.", err)
		return err
	}
	log.Println(reply)

	return nil
}
