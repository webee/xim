package server

import (
	"time"

	"github.com/imdario/mergo"
)

// Config is the configs for http api server.
type Config struct {
	Debug          bool
	Testing        bool
	Key            []byte
	Addr           string
	LogicRPCAddr   string
	RPCCallTimeout time.Duration
}

var (
	defaultConfig = &Config{
		Debug:          false,
		Testing:        false,
		RPCCallTimeout: 5 * time.Second,
	}
)

// NewConfig merge config to default config.
func NewConfig(config *Config) *Config {
	var finalConfig = defaultConfig
	mergo.MergeWithOverwrite(finalConfig, config)
	return finalConfig
}
