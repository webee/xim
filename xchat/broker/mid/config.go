package mid

import (
	"time"

	"github.com/imdario/mergo"
)

// Config is the configs for http api server.
type Config struct {
	Debug          bool
	Testing        bool
	Key            []byte
	LogicRPCAddr   string
	LogicPubAddr   string
	XChatHostURL   string
	RPCCallTimeout time.Duration
}

var (
	defaultConfig = &Config{
		Debug:          false,
		Testing:        false,
		XChatHostURL:   "http://localhost:9980",
		RPCCallTimeout: 5 * time.Second,
	}
)

// NewConfig merge config to default config.
func NewConfig(config *Config) *Config {
	var finalConfig = defaultConfig
	mergo.MergeWithOverwrite(finalConfig, config)
	return finalConfig
}
