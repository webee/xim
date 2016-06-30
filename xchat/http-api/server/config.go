package server

import "time"

// Config is the configs for http api server.
type Config struct {
	Debug          bool
	Testing        bool
	Key            []byte
	Addr           string
	LogicRPCAddr   string
	RPCCallTimeout time.Duration
}
