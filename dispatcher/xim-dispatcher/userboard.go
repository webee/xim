package main

import "xim/broker/userdb"

func initUserboard() {
	userdb.InitUserDB(&userdb.Config{
		RedisNetAddr: args.redisNetAddr,
		Debug:        args.debug,
	})
}
