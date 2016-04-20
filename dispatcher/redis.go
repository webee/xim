package dispatcher

import (
	"time"

	"github.com/garyburd/redigo/redis"
	"gopkg.in/redsync.v1"
)

// NewRedisPool returns a redis pool.
func NewRedisPool(server, password string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     10,
		IdleTimeout: 1 * time.Hour,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", server)
			if err != nil {
				return nil, err
			}
			if len(password) > 0 {
				if _, err = c.Do("AUTH", password); err != nil {
					c.Close()
					return nil, err
				}
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

// NewRedisPools returns many pools.
func NewRedisPools(count int, server, password string) []redsync.Pool {
	pools := make([]redsync.Pool, count)
	for i := 0; i < count; i++ {
		pools = append(pools, NewRedisPool(server, password))
	}
	return pools
}
