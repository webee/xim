package main

import "xim/broker/userdb"

func initUserboard() {
	userdb.InitUserDB(&userdb.Config{
		RedisNetAddr:  args.redisNetAddr,
		RedisPassword: args.redisPassword,
		Debug:         args.debug,
	})
}
