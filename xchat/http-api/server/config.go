package server

import "time"

// Config is the configs for http api server.
type Config struct {
	Debug          bool
	Testing        bool
	Keys           map[string][]byte
	Addr           string
	LogicRPCAddr   string
	RPCCallTimeout time.Duration
	TurnUser       string
	TurnSecret     string
}
