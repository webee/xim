package httpapi

import "github.com/imdario/mergo"

// ServerConfig is the configs for http api server.
type ServerConfig struct {
	Debug       bool
	Addr        string
	SaltPath    string
	AppKeyPath  string
	UserKeyPath string
}

var (
	defaultServerConfig = &ServerConfig{
		Debug: false,
		Addr:  "localhost:6880",
	}
)

// NewServerConfig merge config to default config.
func NewServerConfig(config *ServerConfig) *ServerConfig {
	var finalConfig = defaultServerConfig
	mergo.MergeWithOverwrite(finalConfig, config)
	return finalConfig
}
