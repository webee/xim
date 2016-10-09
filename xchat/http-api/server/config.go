package server

import "time"

// Config is the configs for http api server.
type Config struct {
	Debug           bool
	Testing         bool
	Key             []byte
	Keys            map[string][]byte
	Addr            string
	LogicRPCAddr    string
	RPCCallTimeout  time.Duration
	XChatHostURL    string
	TurnUser        string
	TurnSecret      string
	TurnPasswordTTL int64
	TurnURI         string
}
