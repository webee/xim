package db

import (
	"encoding/json"
	"errors"
	"time"
	"xim/utils/dbutils"
	"xim/utils/netutils"
	"xim/xchat/xpush/mq"
)

const (
	USER_DEV_INFO_KEY = "user:dev:info"
)

var (
	redisConnPool *dbutils.RedisConnPool
)

func InitRedisPool(addr, password string, poolSize int) func() {
	netAddr := &netutils.NetAddr{Network: "tcp", LAddr: addr}
	redisConnPool = dbutils.NewRedisConnPool(netAddr, password, 0, poolSize, 2*poolSize, 30*time.Second)
	return func() {
		if !redisConnPool.IsClosed() {
			redisConnPool.Close()
		}
	}
}

func GetUserDeviceInfo(user string) (*mq.UserDeviceInfo, error) {
	l.Info("GetUserDeviceInfo %s", user)
	ret := &mq.UserDeviceInfo{}

	conn, err := redisConnPool.Get()
	if err != nil {
		l.Error("redisConnPool.Get failed. %v", err)
		return ret, err
	}
	defer redisConnPool.Put(conn)

	reply, err := conn.Do("hget", USER_DEV_INFO_KEY, user)
	if err != nil {
		l.Error("redis.Send failed. %v", err)
		return ret, err
	}

	if reply == nil {
		l.Info("user device info not found. %s", user)
		return ret, errors.New("user device info not found.")
	}

	var data []byte
	switch v := reply.(type) {
	case string:
		data = []byte(v)
	case []byte:
		data = v
	}

	l.Info("GetUserDeviceInfo %s", string(data))

	err = json.Unmarshal(data, &ret)
	if err != nil {
		l.Error("json.Unmarshal failed. %v", err)
		return ret, err
	}

	return ret, nil
}

func SetUserDeviceInfo(user string, udi *mq.UserDeviceInfo) error {
	l.Info("SetUserDeviceInfo %s %v", user, *udi)
	conn, err := redisConnPool.Get()
	if err != nil {
		l.Error("redisConnPool.Get failed. %v", err)
		return err
	}
	defer redisConnPool.Put(conn)

	json, err := json.Marshal(udi)
	if err != nil {
		l.Error("json.Marshal failed. %v", err)
		return err
	}
	reply, err := conn.Do("hset", USER_DEV_INFO_KEY, user, string(json))
	if err != nil {
		l.Error("redis.Send failed. %v", err)
		return err
	}
	l.Info("%v", reply)

	return nil
}
