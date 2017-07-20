package nsutils

import (
	"strings"
)

// EncodeNSUser encode ns user.
func EncodeNSUser(ns, name string) string {
	if ns == "" {
		return name
	}
	return ns + ":" + name
}

// DecodeNSUser decode nsUser.
func DecodeNSUser(user string) (string, string) {
	parts := strings.SplitN(user, ":", 2)
	if len(parts) < 2 {
		return "", user
	}
	return parts[0], parts[1]
}
