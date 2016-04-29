package userboard

// InitUserboard initialize the userboard.
func InitUserboard(c *Config) {
	initConfig(c)
	setupKeys(config)
}
