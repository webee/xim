package mid

import "github.com/imdario/mergo"

// Config is the configs for http api server.
type Config struct {
	Debug        bool
	Testing      bool
	LogicRPCAddr string
	Key          []byte
}

var (
	defaultConfig = &Config{
		Debug:   false,
		Testing: false,
	}
)

// NewConfig merge config to default config.
func NewConfig(config *Config) *Config {
	var finalConfig = defaultConfig
	mergo.MergeWithOverwrite(finalConfig, config)
	return finalConfig
}
