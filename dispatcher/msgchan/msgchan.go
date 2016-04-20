package msgchan

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
	"xim/logic"
)

type queueMsg struct {
	user    logic.UserLocation
	msgType string
	msg     interface{}
	id      chan string
}

func (qm *queueMsg) String() string {
	switch qm.msg.(type) {
	case json.RawMessage:
		return fmt.Sprintf("%s: %s", qm.user, string(qm.msg.(json.RawMessage)))
	default:
		return fmt.Sprintf("%s: %v", qm.user, qm.msg)
	}
}

type chanMsg struct {
	id      string
	user    logic.UserLocation
	msgType string
	msg     interface{}
}

func (cm *chanMsg) String() string {
	switch cm.msg.(type) {
	case json.RawMessage:
		return fmt.Sprintf("%s: %s[%s]", cm.user, string(cm.msg.(json.RawMessage)), cm.id)
	default:
		return fmt.Sprintf("%s: %v", cm.user, cm.msg)
	}
}

// MsgHandler is the function to handle msg.
type MsgHandler func(channel string, user logic.UserLocation, msgType, id, lastID string, msg interface{})

// MsgChan is an abstract message channel.
type MsgChan struct {
	sync.RWMutex
	name       string
	expiration time.Duration
	q          chan *queueMsg
	ch         chan *chanMsg
	closed     bool
	onclose    func()
	quit       chan bool
	count      uint
	idGen      *IDGenerator
	lastID     string
	msgHandler MsgHandler
}

// NewMsgChan creates and starts a msg chan.
func NewMsgChan(name string, expiration time.Duration) *MsgChan {
	msgChan := &MsgChan{
		name:       name,
		expiration: expiration,
		q:          make(chan *queueMsg, 10),
		ch:         make(chan *chanMsg, 100),
		quit:       make(chan bool, 1),
		idGen:      NewIDGenerator(),
	}
	// initial get last id.
	log.Printf("#%s open.", name)
	go msgChan.handleQueue()
	go msgChan.handleChannel()

	return msgChan
}

// Count returns the handled msg count.
func (c *MsgChan) Count() uint {
	return c.count
}

// Close closes the msg channel.
func (c *MsgChan) Close() {
	c.Lock()
	defer c.Unlock()
	if !c.closed {
		close(c.q)
		<-c.quit
		c.closed = true
		log.Printf("#%s closed.", c.name)
		if c.onclose != nil {
			c.onclose()
		}
	}
}

// OnClose set on close action.
func (c *MsgChan) OnClose(f func()) {
	c.onclose = f
}

// SetupMsgHandler setup msg handler.
func (c *MsgChan) SetupMsgHandler(handler MsgHandler) {
	c.msgHandler = handler
}

// Closed returns the channel close status.
func (c *MsgChan) Closed() bool {
	c.RLock()
	defer c.RUnlock()
	return c.closed
}

// Put puts the msg to the msg channel and returns the msg id.
func (c *MsgChan) Put(user logic.UserLocation, msgType string, msg interface{}) (id string, err error) {
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
	// queuing.
	qm := &queueMsg{user, msgType, msg, make(chan string, 1)}
	select {
	case c.q <- qm:
	case <-time.After(1 * time.Second):
		err = errors.New("put timeout")
		return
	}

	select {
	case id = <-qm.id:
	case <-time.After(1 * time.Second):
		err = errors.New("wait timeout")
		return
	}

	return
}

func (c *MsgChan) handleQueue() {
	log.Printf("#%s queue start.", c.name)
	defer func() {
		log.Printf("#%s queue stop.", c.name)
		close(c.ch)
	}()
	for {
		select {
		case qm, more := <-c.q:
			if !more {
				return
			}
			id := c.idGen.ID()
			qm.id <- id
			log.Println("queue:", qm)
			cm := &chanMsg{id, qm.user, qm.msgType, qm.msg}
			c.ch <- cm
		case <-time.After(c.expiration):
			go c.Close()
			continue
		}
	}
}

func (c *MsgChan) handleChannel() {
	log.Printf("#%s channel start.", c.name)
	defer func() {
		log.Printf("#%s channel stop.", c.name)
		c.quit <- true
	}()

	for cm := range c.ch {
		c.count++
		log.Println("channel:", cm)
		// dispatch msg.
		if c.msgHandler != nil {
			c.msgHandler(c.name, cm.user, cm.msgType, cm.id, c.lastID, cm.msg)
		}
		c.lastID = cm.id
	}
}
