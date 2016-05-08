package envutils

import "syscall"

// GetEnvDefault get <key> env with a default value.
func GetEnvDefault(key string, def string) string {
	if val, ok := syscall.Getenv(key); ok {
		return val
	}
	return def
}
