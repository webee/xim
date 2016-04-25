package main

import "xim/broker/userboard"

func initUserboard() {
	userboard.InitUserboard(&userboard.Config{
		RedisNetAddr: args.redisNetAddr,
		UserTimeout:  args.userTimeout,
	})
}
