package userboard

import (
	"log"
	"xim/utils/netutils"
)

// InitUserboard initialize the userboard.
func InitUserboard(c *Config) {
	initConfig(c)

	netAddr, err := netutils.ParseNetAddr(config.RedisNetAddr)
	if err != nil {
		log.Fatalln("bad redis net addr:", config.RedisNetAddr)
	}
	initRedisConnection(netAddr)
}
