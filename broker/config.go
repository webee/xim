package broker

import (
	"time"

	"github.com/imdario/mergo"
)

// WebsocketServerConfig is the configs for websocket server.
type WebsocketServerConfig struct {
	Testing          bool
	Addr             string
	AuthTimeout      time.Duration
	HeartBeatTimeout time.Duration
	WriteTimeout     time.Duration
}

var (
	defaultWebsocketServerConfig = WebsocketServerConfig{
		Testing:          false,
		Addr:             "localhost:2780",
		AuthTimeout:      50 * time.Second,
		HeartBeatTimeout: 120 * time.Second,
		WriteTimeout:     50 * time.Second,
	}
)

// NewWebsocketServerConfig merge config to default config.
func NewWebsocketServerConfig(config *WebsocketServerConfig) *WebsocketServerConfig {
	var finalConfig = defaultWebsocketServerConfig
	mergo.MergeWithOverwrite(&finalConfig, config)
	return &finalConfig
}
