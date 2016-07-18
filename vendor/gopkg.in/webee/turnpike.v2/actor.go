package turnpike

// Actor is a act executor.
type Actor interface {
	Start()
	Close()
	Acts() chan<- ActFunc
	SyncAct(act ActFunc)
}

// ActFunc is actor executable.
type ActFunc func()

// ChannelActor is channel based actor.
type ChannelActor struct {
	acts chan ActFunc
}

// NewChannelActor creates a new channel actor.
func NewChannelActor() Actor {
	return &ChannelActor{
		acts: make(chan ActFunc),
	}
}

func (a *ChannelActor) acting() {
	for act := range a.acts {
		act()
	}
}

// Start starts actor.
func (a *ChannelActor) Start() {
	go a.acting()
}

// Close closes actor.
func (a *ChannelActor) Close() {
	close(a.acts)
}

// Acts returns actor input.
func (a *ChannelActor) Acts() chan<- ActFunc {
	return a.acts
}

// SyncAct execute act synchronized.
func (a *ChannelActor) SyncAct(act ActFunc) {
	sync := make(chan struct{})
	a.acts <- func() {
		act()
		sync <- struct{}{}
	}

	<-sync
}
