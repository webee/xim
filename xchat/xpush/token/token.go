package token

//package main

import (
	"encoding/json"
	"github.com/garyburd/redigo/redis"
	"log"
	"time"
	"xim/xchat/xpush/kafka"
	"errors"
)

const (
	USER_DEV_INFO_KEY = "user:dev:info"
)

func GetUserDeviceInfo(addr, user string) (*kafka.UserDeviceInfo, error) {
	log.Println("GetUserDeviceInfo", addr, user)
	ret := &kafka.UserDeviceInfo{}

	conn, err := redis.Dial("tcp", addr, redis.DialConnectTimeout(30*time.Second),
		redis.DialReadTimeout(10*time.Second), redis.DialWriteTimeout(10*time.Second))
	if err != nil {
		log.Println("redis.Dial failed.", err)
		return ret, err
	}
	defer conn.Close()

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
	conn, err := redis.Dial("tcp", addr, redis.DialConnectTimeout(30*time.Second),
		redis.DialReadTimeout(10*time.Second), redis.DialWriteTimeout(10*time.Second))
	if err != nil {
		log.Println("redis.Dial failed.", err)
		return err
	}
	defer conn.Close()

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

