package mid

import "github.com/imdario/mergo"

// Config is the configs for http api server.
type Config struct {
	Debug        bool
	Testing      bool
	XIMHostURL   string
	XIMApp       string
	XIMPassword  string
	XIMAppWsURL  string
	Key          []byte
	XChatHostURL string
}

var (
	defaultConfig = &Config{
		Debug:        false,
		Testing:      false,
		XIMHostURL:   "http://localhost:6980",
		XIMApp:       "test",
		XIMPassword:  "test1234",
		XIMAppWsURL:  "ws://127.0.0.1:2980/ws",
		XChatHostURL: "http://localhost:9980",
	}
)

// NewConfig merge config to default config.
func NewConfig(config *Config) *Config {
	var finalConfig = defaultConfig
	mergo.MergeWithOverwrite(finalConfig, config)
	return finalConfig
}
