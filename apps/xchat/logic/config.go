package logic

import "github.com/imdario/mergo"

// Config is the configs for http api server.
type Config struct {
	Debug     bool
	BrokerURL string
}

var (
	defaultConfig = &Config{
		Debug:     false,
		BrokerURL: "ws://127.0.0.1:48079/app-ws",
	}
)

// NewConfig merge config to default config.
func NewConfig(config *Config) *Config {
	var finalConfig = defaultConfig
	mergo.MergeWithOverwrite(finalConfig, config)
	return finalConfig
}
