package rpcservice

import (
	"log"
	"time"
	"xim/dispatcher/msgchan"
	"xim/utils/syncutils"

	"github.com/webee/ttlcache"
)

type msgChannelCache struct {
	ko    *syncutils.KeyOnce
	cache *ttlcache.Cache
	new   func(key string) *msgchan.MsgChannel
}

var (
	channelCache     = newMsgChannelCache(60*time.Second, newDispatcherMsgChan)
	userChannelCache = newMsgChannelCache(65*time.Second, newUserMsgChan)
)

func newMsgChannelCache(ttl time.Duration, new func(key string) *msgchan.MsgChannel) *msgChannelCache {
	c := &msgChannelCache{
		ko:    syncutils.NewKeyOnce("msg.channel.cache", 1*time.Hour),
		cache: ttlcache.NewCache(),
		new:   new,
	}

	c.cache.SetTTL(ttl)
	c.cache.SetCheckExpirationCallback(func(key string, value interface{}) bool {
		msgChan := value.(*msgchan.MsgChannel)
		c.ko.UndoOnKey(key)
		msgChan.Close()
		return msgChan.Closed()
	})
	c.cache.SetExpirationCallback(func(key string, value interface{}) {
		log.Printf("#%s MsgChannel expired.", key)
	})
	return c
}

func (c *msgChannelCache) getMsgChan(key string) *msgchan.MsgChannel {
	item, ok := c.cache.Get(key)
	if !ok {
		c.ko.DoOnKey(key, func() {
			item = c.new(key)
			c.cache.Set(key, item)
		})
		item, _ = c.cache.Get(key)
	}

	return item.(*msgchan.MsgChannel)
}
