package main

import (
	"xim/broker/userboard"
	"xim/broker/userdb"
)

func initUserboard() {
	userdb.InitUserDB(&userdb.Config{
		RedisNetAddr: args.redisNetAddr,
		UserTimeout:  args.userTimeout,
		Debug:        args.debug,
	})
	userboard.InitUserboard(&userboard.Config{
		AppKeyPath:  args.appKeyPath,
		UserKeyPath: args.userKeyPath,
		Debug:       args.debug,
	})
}
