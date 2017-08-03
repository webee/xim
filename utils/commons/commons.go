package commons

import "strings"

// TrimBytesSpace trim bytes spaces.
func TrimBytesSpace(b []byte) []byte {
	return []byte(strings.TrimSpace(string(b)))
}
