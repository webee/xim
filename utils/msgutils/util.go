package msgutils

import (
	"math/rand"
	"time"
)

const (
	maxID int64 = 1 << 53
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// NewID generates a random ID.
func NewID() ID {
	return ID(rand.Int63n(maxID))
}
