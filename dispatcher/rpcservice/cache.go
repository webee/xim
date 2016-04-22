package rpcservice

import (
	"log"
	"sync"
	"time"
	"xim/dispatcher/msgchan"

	"github.com/webee/ttlcache"
)

type msgChannelCache struct {
	sync.RWMutex
	cache *ttlcache.Cache
}

var (
	channelCache = newMsgChannelCache()
)

func newMsgChannelCache() *msgChannelCache {
	c := &msgChannelCache{
		cache: ttlcache.NewCache(),
	}

	c.cache.SetTTL(10 * time.Second)
	c.cache.SetCheckExpirationCallback(func(key string, value interface{}) bool {
		msgChan := value.(*msgchan.MsgChannel)
		msgChan.Close()
		return true
	})
	c.cache.SetExpirationCallback(func(key string, value interface{}) {
		log.Printf("#%s MsgChannel expired.", key)
	})
	return c
}

func (c *msgChannelCache) getMsgChan(channel string) (msgChan *msgchan.MsgChannel) {
	item, exists := c.cache.Get(channel)
	if exists && !item.(*msgchan.MsgChannel).Closed() {
		return item.(*msgchan.MsgChannel)
	}

	c.Lock()
	defer c.Unlock()
	if item, exists = c.cache.Get(channel); !exists || item.(*msgchan.MsgChannel).Closed() {
		msgChan = newDispatcherMsgChan(channel)
		c.cache.Set(channel, msgChan)
	}
	return
}
