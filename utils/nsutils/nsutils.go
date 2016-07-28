package nsutils

import (
	"strings"
)

// EncodeNSUser encode ns user.
func EncodeNSUser(ns, u string) (user string) {
	if ns == "" {
		return u
	}
	return ns + ":" + u
}

// DecodeNSUser decode nsUser.
func DecodeNSUser(user string) (string, string) {
	parts := strings.SplitN(user, ":", 2)
	if len(parts) < 2 {
		return "", user
	}
	return parts[0], parts[1]
}
