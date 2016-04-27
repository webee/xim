package ttlcache

import (
	"sync"
	"time"
)

// CheckExpireCallback is used as a callback for an external check on item expiration
type checkExpireCallback func(key string, value interface{}) bool

// ExpireCallback is used as a callback on item expiration or when notifying of an item new to the cache
type expireCallback func(key string, value interface{})

// Cache is a synchronized map of items that can auto-expire once stale
type Cache struct {
	sync.RWMutex
	ttl                    time.Duration
	items                  map[string]*item
	expireCallback         expireCallback
	checkExpireCallback    checkExpireCallback
	newItemCallback        expireCallback
	priorityQueue          *priorityQueue
	expirationNotification chan bool
	expirationTime         time.Time
}

func (cache *Cache) getItem(key string) (*item, bool) {
	cache.RLock()
	defer cache.RUnlock()

	item, exists := cache.items[key]
	if !exists || item.expired() {
		return nil, false
	}

	if item.ttl >= 0 && (item.ttl > 0 || cache.ttl > 0) {
		if cache.ttl > 0 && item.ttl == 0 {
			item.ttl = cache.ttl
		}

		item.touch()
		cache.priorityQueue.update(item)
		cache.expirationNotificationTrigger(item)
	}

	return item, exists
}

func (cache *Cache) startExpirationProcessing() {
	for {
		var sleepTime time.Duration
		cache.Lock()
		if cache.priorityQueue.Len() > 0 {
			if cache.ttl > 0 && time.Now().Add(cache.ttl).Before(cache.priorityQueue.items[0].expireAt) {
				sleepTime = cache.ttl
			} else {
				sleepTime = cache.priorityQueue.items[0].expireAt.Sub(time.Now())
			}
		} else if cache.ttl > 0 {
			sleepTime = cache.ttl
		} else {
			sleepTime = time.Duration(1 * time.Hour)
		}

		cache.expirationTime = time.Now().Add(sleepTime)
		cache.Unlock()

		select {
		case <-time.After(cache.expirationTime.Sub(time.Now())):
			if cache.priorityQueue.Len() == 0 {
				continue
			}

			cache.Lock()
			item := cache.priorityQueue.items[0]

			if !item.expired() {
				cache.Unlock()
				continue
			}

			if cache.checkExpireCallback != nil {
				if !cache.checkExpireCallback(item.key, item.data) {
					item.touch()
					cache.priorityQueue.update(item)
					cache.Unlock()
					continue
				}
			}

			cache.priorityQueue.remove(item)
			delete(cache.items, item.key)
			cache.Unlock()

			if cache.expireCallback != nil {
				cache.expireCallback(item.key, item.data)
			}
		case <-cache.expirationNotification:
			continue
		}
	}
}

func (cache *Cache) expirationNotificationTrigger(item *item) {
	if cache.expirationTime.After(time.Now().Add(item.ttl)) {
		cache.expirationNotification <- true
	}
}

// Set is a thread-safe way to add new items to the map
func (cache *Cache) Set(key string, data interface{}) {
	cache.SetWithTTL(key, data, ItemExpireWithGlobalTTL)
}

// SetWithTTL is a thread-safe way to add new items to the map with individual ttl
func (cache *Cache) SetWithTTL(key string, data interface{}, ttl time.Duration) {
	item, exists := cache.getItem(key)
	cache.Lock()

	if exists {
		item.data = data
		item.ttl = ttl
	} else {
		item = newItem(key, data, ttl)
		cache.items[key] = item
	}

	if item.ttl >= 0 && (item.ttl > 0 || cache.ttl > 0) {
		if cache.ttl > 0 && item.ttl == 0 {
			item.ttl = cache.ttl
		}

		item.touch()

		if exists {
			cache.priorityQueue.update(item)
		} else {
			cache.priorityQueue.push(item)
		}
		cache.expirationNotificationTrigger(item)
	}
	cache.Unlock()
	if !exists && cache.newItemCallback != nil {
		cache.newItemCallback(key, data)
	}
}

// Get is a thread-safe way to lookup items
// Every lookup, also touches the item, hence extending it's life
func (cache *Cache) Get(key string) (interface{}, bool) {
	item, exists := cache.getItem(key)
	if exists {
		return item.data, true
	}
	return nil, false
}

// Remove remove a key.
func (cache *Cache) Remove(key string) bool {
	cache.Lock()
	defer cache.Unlock()
	object, exists := cache.items[key]
	if !exists {
		return false
	}
	delete(cache.items, object.key)
	cache.priorityQueue.remove(object)

	return true
}

// Count returns the number of items in the cache
func (cache *Cache) Count() int {
	cache.RLock()
	defer cache.RUnlock()
	length := len(cache.items)
	return length
}

// SetTTL set the global ttl.
func (cache *Cache) SetTTL(ttl time.Duration) {
	cache.Lock()
	defer cache.Unlock()
	cache.ttl = ttl
	cache.expirationNotification <- true
}

// SetExpirationCallback sets a callback that will be called when an item expires
func (cache *Cache) SetExpirationCallback(callback expireCallback) {
	cache.expireCallback = callback
}

// SetCheckExpirationCallback sets a callback that will be called when an item is about to expire
// in order to allow external code to decide whether the item expires or remains for another TTL cycle
func (cache *Cache) SetCheckExpirationCallback(callback checkExpireCallback) {
	cache.checkExpireCallback = callback
}

// SetNewItemCallback sets a callback that will be called when a new item is added to the cache
func (cache *Cache) SetNewItemCallback(callback expireCallback) {
	cache.newItemCallback = callback
}

// NewCache is a helper to create instance of the Cache struct
func NewCache() *Cache {
	cache := &Cache{
		items:                  make(map[string]*item),
		priorityQueue:          newPriorityQueue(),
		expirationNotification: make(chan bool, 1),
		expirationTime:         time.Now(),
	}
	go cache.startExpirationProcessing()
	return cache
}
