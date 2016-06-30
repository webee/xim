package mid

import "time"

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
