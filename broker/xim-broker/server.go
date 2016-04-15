package main

import (
	"fmt"
	"log"
	"time"

	"gopkg.in/go-redis/cache.v3"
	"gopkg.in/redis.v3"
	"gopkg.in/vmihailenco/msgpack.v2"
)

// ServerStatus represents server status.
type ServerStatus struct {
	ID     int
	status string
	tick   int
}

var (
	serverStatus = ServerStatus{}
	statusChan   = make(chan bool)

	redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	redisCache = &cache.Codec{
		Redis: redisClient,
		Marshal: func(v interface{}) ([]byte, error) {
			return msgpack.Marshal(v)
		},
		Unmarshal: func(b []byte, v interface{}) error {
			return msgpack.Unmarshal(b, v)
		},
	}
)

func updatingStatus() {
	for {
		select {
		case <-statusChan:
			log.Println("status: ", serverStatus)
			if serverStatus.ID != 0 {
				err := redisCache.Set(&cache.Item{
					Key:        serverStatus.Key(),
					Object:     serverStatus,
					Expiration: 14 * time.Second,
				})
				if err != nil {
					log.Panicln(err)
				}
				serverStatus.tick++
			}
			go (func() {
				time.Sleep(10 * time.Second)
				statusChan <- true
			})()
		}
	}
}

func setupServer() {
	// continuing updating server status.
	go updatingStatus()

	var (
		id             int
		serverStatuses = make(map[int]ServerStatus)
	)

	serverKeys, err := redisClient.Keys("xim:broker:*").Result()
	if err != nil {
		log.Panicln(err)
	}
	for _, serverKey := range serverKeys {
		var status ServerStatus
		if err := redisCache.Get(serverKey, &serverStatus); err != nil {
			log.Panicln(err)
		}
		serverStatuses[serverStatus.ID] = status
	}
	for id = 1; id <= 1000; id++ {
		if _, ok := serverStatuses[id]; !ok {
			break
		}
	}
	serverStatus.ID = id
	serverStatus.SettingStatus("initing")
}

// Key returns the cache key.
func (s *ServerStatus) Key() string {
	return fmt.Sprintf("xim:broker:%d", s.ID)
}

// SettingStatus setting and updating server status.
func (s *ServerStatus) SettingStatus(status string) {
	s.status = status
	statusChan <- true
}
