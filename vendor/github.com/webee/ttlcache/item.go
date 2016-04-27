package ttlcache

import (
	"sync"
	"time"
)

const (
	ItemNotExpire           time.Duration = -1
	ItemExpireWithGlobalTTL time.Duration = 0
)

func newItem(key string, data interface{}, ttl time.Duration) *item {
	item := &item{
		data: data,
		ttl:  ttl,
		key:  key,
	}
	item.touch()
	return item
}

type item struct {
	sync.RWMutex
	key        string
	data       interface{}
	ttl        time.Duration
	expireAt   time.Time
	queueIndex int
}

// Reset the item expiration time
func (item *item) touch() {
	item.Lock()
	defer item.Unlock()
	if item.ttl > 0 {
		item.expireAt = time.Now().Add(item.ttl)
	}
}

// Verify if the item is expired
func (item *item) expired() bool {
	item.RLock()
	defer item.RUnlock()
	if item.ttl <= 0 {
		return false
	}
	expired := item.expireAt.Before(time.Now())
	return expired
}
