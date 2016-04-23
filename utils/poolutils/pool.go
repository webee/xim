package poolutils

import (
	"sync"
)

const (
	badID = -1
)

// Resource is a closable object.
type Resource interface {
	Close()
}

type resourceCounter struct {
	res   Resource
	count uint
}

// ConcurResPool is a resource pool for resource that can be used concurrently.
type ConcurResPool struct {
	sync.RWMutex
	resoures    []*resourceCounter
	new         func() (Resource, error)
	concurrency uint
}

// NewConcurObjPool creates a new ConcurResPool.
func NewConcurObjPool(concurrency uint, new func() (Resource, error)) *ConcurResPool {
	return &ConcurResPool{
		resoures:    []*resourceCounter{},
		new:         new,
		concurrency: concurrency,
	}
}

// Get returns a resource.
func (p *ConcurResPool) Get() (id int, res Resource, err error) {
	p.Lock()
	defer p.Unlock()
	var resCnt *resourceCounter
	for id, resCnt = range p.resoures {
		if resCnt.count < p.concurrency {
			resCnt.count++
			return id, resCnt.res, nil
		}
	}

	res, err = p.new()
	if err != nil {
		return badID, res, err
	}
	id = len(p.resoures)
	p.resoures = append(p.resoures, &resourceCounter{
		res:   res,
		count: 1,
	})
	return id, res, err
}

// Put put a resource back.
func (p *ConcurResPool) Put(id int) {
	if id == badID {
		return
	}

	p.Lock()
	p.resoures[id].count--
	p.Unlock()
}
