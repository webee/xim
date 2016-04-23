package syncutils

import (
	"log"
	"sync"
	"time"
)

// KeyOnce perform action for key only once.
type KeyOnce struct {
	sync.RWMutex
	name   string
	keymap map[string]*Once
	ticker *time.Ticker
}

const (
	// NoCleanup is the cleanupPeriod constant represents no cleanup.
	NoCleanup = 0
)

// NewKeyOnce creates a KeyOnce.
func NewKeyOnce(name string, cleanupPeriod time.Duration) *KeyOnce {
	ko := &KeyOnce{
		name:   name,
		keymap: make(map[string]*Once),
	}
	if cleanupPeriod > NoCleanup {
		ko.ticker = time.NewTicker(cleanupPeriod)
		go ko.cleanup()
	}
	return ko
}

// DoOnKey call f for key only once.
func (ko *KeyOnce) DoOnKey(key string, f func()) {
	ko.RLock()
	once, ok := ko.keymap[key]
	if !ok {
		ko.RUnlock()
		ko.Lock()
		if once, ok = ko.keymap[key]; !ok {
			once = &Once{}
			ko.keymap[key] = once
		}
		ko.Unlock()
	} else {
		ko.RUnlock()
	}

	once.Do(f)
}

// UndoOnKey reset key once.
func (ko *KeyOnce) UndoOnKey(key string) {
	ko.RLock()
	defer ko.RUnlock()
	if once, ok := ko.keymap[key]; ok {
		once.Undo()
	}
}

func (ko *KeyOnce) cleanup() {
	for _ = range ko.ticker.C {
		log.Printf("KeyOnce[%s] Cleanup\n", ko.name)
		ko.Lock()
		ks := []string{}
		for k, v := range ko.keymap {
			if v.Done() {
				ks = append(ks, k)
			}
		}
		for _, k := range ks {
			log.Printf("KeyOnce[%s] delete %s\n", ko.name, k)
			delete(ko.keymap, k)
		}
		ko.Unlock()
	}
}

// Close stop ticker.
func (ko *KeyOnce) Close() {
	if ko.ticker != nil {
		ko.ticker.Stop()
	}
}
