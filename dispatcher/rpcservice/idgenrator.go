package rpcservice

// IDGenerator is a time based id generator.
type IDGenerator struct {
	id uint64
}

// NewIDGenerator creates a new id generator.
func NewIDGenerator() *IDGenerator {
	return &IDGenerator{}
}

// SetID sets the id.
func (g *IDGenerator) SetID(id uint64) {
	g.id = id
}

// ID return the next id.
func (g *IDGenerator) ID() uint64 {
	g.id++
	return g.id
}
