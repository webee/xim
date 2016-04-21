package rpcservice

import (
	"log"
	"sync"
	"time"
	"xim/dispatcher/msgchan"
)

type channelCache struct {
	sync.RWMutex
	channels map[string]*msgchan.MsgChannel
}

var (
	channels = &channelCache{
		channels: make(map[string]*msgchan.MsgChannel, 1000),
	}
)

func (c *channelCache) getMsgChan(channel string) (msgChan *msgchan.MsgChannel) {
	c.RLock()
	msgChan, ok := c.channels[channel]
	c.RUnlock()
	if !ok || msgChan.Closed() {
		c.Lock()
		if msgChan, ok = c.channels[channel]; !ok || msgChan.Closed() {
			msgChan = newDispatcherMsgChan(channel, 10*time.Second).OnClose(func() {
				c.Lock()
				defer c.Unlock()
				delete(c.channels, channel)
				log.Printf("delete #%s from channels.", channel)
			})
			c.channels[channel] = msgChan
		}
		c.Unlock()
	}
	return
}
