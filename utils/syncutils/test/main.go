package main

import (
	"log"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"
	"xim/utils/syncutils"
)

type Cache struct {
	sync.RWMutex
	ko *syncutils.KeyOnce
	kv map[string]int
}

func NewCache() *Cache {
	return &Cache{
		ko: syncutils.NewKeyOnce("cache", 2*time.Second),
		kv: make(map[string]int),
	}
}

func (c *Cache) Get(key string) int {
	parts := strings.Split(key, ".")
	key = parts[0]
	i, _ := strconv.Atoi(parts[1])
	c.ko.DoOnKey(key, func() {
		log.Println(key, i)
		time.Sleep(50 * time.Millisecond)
		c.Lock()
		c.kv[key] = i
		c.Unlock()
	})
	c.RLock()
	defer c.RUnlock()
	return c.kv[key]
}

func (c *Cache) Get2(key string) int {
	parts := strings.Split(key, ".")
	key = parts[0]
	i, _ := strconv.Atoi(parts[1])

	c.RLock()
	v, ok := c.kv[key]
	if ok {
		c.RUnlock()
		return v
	}
	c.RUnlock()

	c.Lock()
	defer c.Unlock()
	if v, ok = c.kv[key]; ok {
		return v
	}

	log.Println(key, i)
	time.Sleep(1 * time.Second)
	c.kv[key] = i
	return i
}

func main() {
	wg := &sync.WaitGroup{}
	c := NewCache()
	keys := []string{
		"a.1", "a.2", "a.3", "a.4",
		"b.1", "b.2", "b.3", "b.4",
		"c.1", "c.2", "c.3", "c.4",
		"d.1", "d.2", "d.3", "d.4",
		"e.1", "e.2", "e.3", "e.4",
		"f.1", "f.2", "f.3", "f.4",
	}

	wg.Add(len(keys))
	for i, k := range keys {
		go func(i int, k string) {
			time.Sleep(time.Duration(rand.Intn(3)) * time.Second)
			log.Println("res:", i, k, c.Get(k))
			wg.Done()
		}(i, k)
	}
	wg.Wait()
}
