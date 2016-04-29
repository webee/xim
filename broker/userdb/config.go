package userdb

import "github.com/imdario/mergo"

// Config is the configs for userboard.
type Config struct {
	RedisNetAddr string
	UserTimeout  int
	Debug        bool
}

var (
	defaultConfig = &Config{
		RedisNetAddr: "tcp@localhost:6379",
		UserTimeout:  12,
		Debug:        false,
	}
	config *Config
)

func initConfig(c *Config) {
	config = &Config{}
	mergo.MergeWithOverwrite(config, defaultConfig)
	mergo.MergeWithOverwrite(config, c)
}
