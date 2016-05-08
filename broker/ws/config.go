package ws

import (
	"time"

	"github.com/imdario/mergo"
)

// WebsocketServerConfig is the configs for websocket server.
type WebsocketServerConfig struct {
	Testing          bool
	Addr             string
	HTTPReadTimeout  time.Duration
	HTTPWriteTimeout time.Duration
	HeartbeatTimeout time.Duration
	WriteTimeout     time.Duration
	Broker           string
}

var (
	defaultWebsocketServerConfig = WebsocketServerConfig{
		Testing:          false,
		Addr:             "localhost:2880",
		HTTPReadTimeout:  7 * time.Second,
		HTTPWriteTimeout: 7 * time.Second,
		HeartbeatTimeout: 12 * time.Second,
		WriteTimeout:     7 * time.Second,
	}
)

// NewWebsocketServerConfig merge config to default config.
func NewWebsocketServerConfig(config *WebsocketServerConfig) *WebsocketServerConfig {
	var finalConfig = defaultWebsocketServerConfig
	mergo.MergeWithOverwrite(&finalConfig, config)
	return &finalConfig
}
