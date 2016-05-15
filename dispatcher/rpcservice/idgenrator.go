package rpcservice

// IDGenerator is a time based id generator.
type IDGenerator struct {
	id int
}

// NewIDGenerator creates a new id generator.
func NewIDGenerator() *IDGenerator {
	return &IDGenerator{}
}

// SetID sets the id.
func (g *IDGenerator) SetID(id int) {
	g.id = id
}

// ID return the next id.
func (g *IDGenerator) ID() int {
	g.id++
	return g.id
}
