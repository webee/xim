package broker

import (
	"xim/utils/netutils"
	"xim/utils/syncutils"

	"github.com/webee/ttlcache"
)

// RPCClientPoolCache is a cache for rpcClientPool.
type RPCClientPoolCache struct {
	ko    *syncutils.KeyOnce
	cache *ttlcache.Cache
	new   func(key string) *rpcClientPool
}

var (
	rpcClientPoolCache = newRPCClientCache(newRPCClientPoolFromKey)
)

func newRPCClientPoolFromKey(key string) *rpcClientPool {
	netAddr, _ := netutils.ParseNetAddr(key)
	pool := newRPCClientPool(netAddr, 100)
	return pool
}

func newRPCClientCache(new func(key string) *rpcClientPool) *RPCClientPoolCache {
	c := &RPCClientPoolCache{
		ko:    syncutils.NewKeyOnce("broker.client.cache", syncutils.NoCleanup),
		cache: ttlcache.NewCache(),
		new:   new,
	}
	return c
}

func (c *RPCClientPoolCache) getRPCClientPool(key string) *rpcClientPool {
	c.ko.DoOnKey(key, func() {
		pool := c.new(key)
		c.cache.Set(key, pool)
	})

	item, _ := c.cache.Get(key)
	return item.(*rpcClientPool)
}
