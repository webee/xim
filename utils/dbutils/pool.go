package dbutils

import (
	"time"

	"xim/utils/netutils"

	"github.com/garyburd/redigo/redis"
	"github.com/youtube/vitess/go/pools"
	"golang.org/x/net/context"
)

// RedisConn adapts a Redigo connection to a Vitess Resource.
type RedisConn struct {
	redis.Conn
}

// Close close the redigo connection.
func (c *RedisConn) Close() {
	c.Conn.Close()
}

// RedisConnPool is a redis connection pool.
type RedisConnPool struct {
	*pools.ResourcePool
}

// NewRedisConnPool creates a redis connection pool.
func NewRedisConnPool(netAddr *netutils.NetAddr, password string, db, capacity, maxCap int, idleTimeout time.Duration) *RedisConnPool {
	p := pools.NewResourcePool(func() (pools.Resource, error) {
		c, err := redis.Dial(netAddr.Network, netAddr.LAddr, redis.DialPassword(password))
		if err != nil {
			return nil, err
		}
		conn := &RedisConn{c}
		_, err = conn.Do("SELECT", db)
		if err != nil {
			conn.Close()
			return nil, err
		}
		return conn, err
	}, capacity, maxCap, idleTimeout)
	return &RedisConnPool{
		p,
	}
}

// Get returns a idle redis connection.
func (p *RedisConnPool) Get() (*RedisConn, error) {
	ctx := context.TODO()
	r, err := p.ResourcePool.Get(ctx)
	if err != nil {
		return nil, err
	}
	return r.(*RedisConn), nil
}

// Put return back the redis connection.
func (p *RedisConnPool) Put(c *RedisConn) {
	p.ResourcePool.Put(c)
}
