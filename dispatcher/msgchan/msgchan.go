package msgchan

import (
	"errors"
	"log"
	"sync"
	"time"
)

// MsgChannelTransformer is the transformer channel msg.
type MsgChannelTransformer func(interface{}) interface{}

// MsgChannelHandler is the handler for msg channel.
type MsgChannelHandler func(interface{}) error

// MsgChannelDownStream is the down stream of this channel.
type MsgChannelDownStream interface {
	Put(interface{}) error
	Close()
}

// MsgChannel is an abstract message channel.
type MsgChannel struct {
	sync.RWMutex
	name        string
	in          chan interface{}
	transformer MsgChannelTransformer
	next        MsgChannelDownStream
	expire      time.Duration
	putTimeout  time.Duration
	closed      bool
	onclose     func()
	quit        chan bool
}

// NewMsgChannelHandlerDownStream creates a msg channel down stream by a handler.
func NewMsgChannelHandlerDownStream(name string, msgChannelHandler MsgChannelHandler) MsgChannelDownStream {
	return &msgChannelHandlerDownStream{
		name:              name,
		msgChannelHandler: msgChannelHandler,
	}
}

// NewMsgChannel creates and starts a msg chan.
func NewMsgChannel(name string, size int, handler MsgChannelTransformer, next MsgChannelDownStream, expire time.Duration) *MsgChannel {
	msgChan := &MsgChannel{
		name:        name,
		in:          make(chan interface{}, size),
		transformer: handler,
		next:        next,
		quit:        make(chan bool, 1),
		expire:      expire,
		putTimeout:  5 * time.Second,
	}
	go msgChan.open()

	return msgChan
}

// Close closes the msg channel.
func (c *MsgChannel) Close() {
	c.Lock()
	defer c.Unlock()
	if !c.closed {
		close(c.in)
		<-c.quit
		c.next.Close()
		c.closed = true
		log.Printf("[%s] closed.", c.name)
		if c.onclose != nil {
			c.onclose()
		}
	}
}

// Put puts the msg in.
func (c *MsgChannel) Put(qm interface{}) (err error) {
	if c.closed {
		err = errors.New("channel closed")
		return
	}

	defer func() {
		if r := recover(); r != nil {
			log.Println(r)
			err = errors.New("channel closed")
		}
	}()

	select {
	case c.in <- qm:
	case <-time.After(c.putTimeout):
		err = errors.New("put timeout")
		return
	}
	return
}

func (c *MsgChannel) open() {
	log.Printf("[%s] open.", c.name)
	log.Printf("[%s] start.", c.name)
	defer func() {
		log.Printf("[%s] stop.", c.name)
		c.quit <- true
	}()
	for {
		select {
		case mIn, more := <-c.in:
			if !more {
				return
			}
			if err := c.next.Put(c.transformer(mIn)); err == nil {
			} else {
				log.Println(err)
			}
		case <-time.After(c.expire):
			go c.Close()
			continue
		}
	}
}

// OnClose set on close action.
func (c *MsgChannel) OnClose(f func()) *MsgChannel {
	c.onclose = f
	return c
}

func (c *MsgChannel) setPutTimeout(t time.Duration) *MsgChannel {
	c.putTimeout = t
	return c
}

// Closed returns the channel close status.
func (c *MsgChannel) Closed() bool {
	c.RLock()
	defer c.RUnlock()
	return c.closed
}

// MsgChannelHandlerDownStream is the down stream of a handler.
type msgChannelHandlerDownStream struct {
	name              string
	msgChannelHandler MsgChannelHandler
}

func (c *msgChannelHandlerDownStream) Put(m interface{}) error {
	return c.msgChannelHandler(m)
}

func (c *msgChannelHandlerDownStream) Close() {
	log.Printf("[%s] finished.", c.name)
}
