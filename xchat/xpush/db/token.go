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
	userDevInfoKey = "user:dev:info"
)

var (
	redisConnPool *dbutils.RedisConnPool
)

// InitRedisPool initialize the redis pool
func InitRedisPool(addr, password string, poolSize int) func() {
	netAddr := &netutils.NetAddr{Network: "tcp", LAddr: addr}
	redisConnPool = dbutils.NewRedisConnPool(netAddr, password, 0, poolSize, 2*poolSize, 30*time.Second)
	return func() {
		if !redisConnPool.IsClosed() {
			redisConnPool.Close()
		}
	}
}

// GetUserDeviceInfo get user's device info
func GetUserDeviceInfo(user string) (*mq.UserDeviceInfo, error) {
	l.Info("GetUserDeviceInfo %s", user)
	ret := &mq.UserDeviceInfo{}

	conn, err := redisConnPool.Get()
	if err != nil {
		l.Warning("redisConnPool.Get failed. %s", err.Error())
		return ret, err
	}
	defer redisConnPool.Put(conn)

	reply, err := conn.Do("HGET", userDevInfoKey, user)
	if err != nil {
		l.Warning("redis.Send failed. %s", err.Error())
		return ret, err
	}

	if reply == nil {
		l.Info("user device info not found. %s", user)
		return ret, errors.New("user device info not found")
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
		l.Warning("json.Unmarshal failed. %s", err.Error())
		return ret, err
	}

	return ret, nil
}

// SetUserDeviceInfo set user's device info
func SetUserDeviceInfo(user string, udi *mq.UserDeviceInfo) error {
	l.Info("SetUserDeviceInfo %s %v", user, *udi)
	conn, err := redisConnPool.Get()
	if err != nil {
		l.Warning("redisConnPool.Get failed. %s", err.Error())
		return err
	}
	defer redisConnPool.Put(conn)

	json, err := json.Marshal(udi)
	if err != nil {
		l.Warning("json.Marshal failed. %s", err.Error())
		return err
	}
	reply, err := conn.Do("HSET", userDevInfoKey, user, string(json))
	if err != nil {
		l.Warning("redis.Send failed. %s", err.Error())
		return err
	}
	l.Info("%v", reply)

	return nil
}
