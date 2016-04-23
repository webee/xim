package rpcservice

import (
	"fmt"
	"time"
)

// IDGenerator is a time based id generator.
type IDGenerator struct {
	ts    int64
	count int
}

// NewIDGenerator creates a new id generator.
func NewIDGenerator() *IDGenerator {
	return &IDGenerator{}
}

// ID return the next id.
func (g *IDGenerator) ID() string {
	ts := time.Now().Unix()
	if ts > g.ts {
		g.ts = ts
		g.count = 0
	}
	g.count++
	return fmt.Sprintf("%d.%06d", g.ts, g.count)
}
