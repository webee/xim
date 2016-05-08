package userboard

import "github.com/imdario/mergo"

// Config is the configs for userboard.
type Config struct {
	AppKeyPath  string
	UserKeyPath string
	Debug       bool
}

var (
	defaultConfig = &Config{
		Debug: false,
	}
	config *Config
)

func initConfig(c *Config) {
	config = &Config{}
	mergo.MergeWithOverwrite(config, defaultConfig)
	mergo.MergeWithOverwrite(config, c)
}
