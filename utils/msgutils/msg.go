package msgutils

import (
	"errors"
	"fmt"
	"log"
	"time"
)

// An ID is a unique, non-negative number.
type ID uint64

// MessageType type.
type MessageType int

// Message is the send/recv message.
type Message interface {
	MessageType() MessageType
}

// SyncMessage is a sync send/reply message.
type SyncMessage interface {
	Message
	SetID(ID)
	GetID() ID
}

// MessageHandler handles selected messages.
type MessageHandler func(msg Message)

// A MsgController controll msg from/to a Peer backend.
type MsgController struct {
	Transeiver
	// ReplyTimeout is the amount of time that the controller will block waiting for a reply from the Transeiver.
	ReplyTimeout  time.Duration
	handler       MessageHandler
	syncListeners map[ID]chan SyncMessage
	acts          chan func()
	closed        bool
	noSync        bool
}

// NewMsgController creates a Transeiver msg controller.
func NewMsgController(t Transeiver, handler MessageHandler) *MsgController {
	return &MsgController{
		Transeiver:    t,
		ReplyTimeout:  10 * time.Second,
		handler:       handler,
		syncListeners: make(map[ID]chan SyncMessage),
		acts:          make(chan func()),
	}
}

// NewNoSyncMsgController creates a Transeiver msg controller without sync send.
func NewNoSyncMsgController(t Transeiver, handler MessageHandler) *MsgController {
	c := &MsgController{
		Transeiver:    t,
		ReplyTimeout:  10 * time.Second,
		handler:       handler,
		syncListeners: make(map[ID]chan SyncMessage),
		acts:          make(chan func()),
		noSync:        true,
	}
	go c.receive()
	return c
}

// Start starts run and receive.
func (c *MsgController) Start() {
	go c.run()
	go c.receive()
}

func (c *MsgController) run() {
	for {
		if act, ok := <-c.acts; ok {
			act()
		} else {
			break
		}
	}
	log.Println("client closed")
}

// Close closes the connection to the server.
func (c *MsgController) Close() error {
	if !c.closed {
		if err := c.Transeiver.Close(); err != nil {
			return fmt.Errorf("error closing Transeiver: %v", err)
		}
		c.closed = true
	}
	return nil
}

// Receive handles messages from the Transeiver until it is disconnected.
func (c *MsgController) receive() {
	for msg := range c.Transeiver.Receive() {
		if c.noSync {
			c.handler(msg)
			continue
		}

		switch msg := msg.(type) {
		case SyncMessage:
			c.notifySyncListener(msg)
		default:
			c.handler(msg)
		}
	}

	close(c.acts)
	c.Close()
}

// SyncSend send msg and wait for reply.
func (c *MsgController) SyncSend(msg SyncMessage) (SyncMessage, error) {
	if c.noSync {
		return nil, errors.New("message controller is no sync")
	}

	id := NewID()
	c.setSyncListener(id)
	msg.SetID(id)
	err := c.Send(msg)
	if err != nil {
		return nil, err
	}
	reply, err := c.waitOnSyncListener(id)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

func (c *MsgController) setSyncListener(id ID) {
	log.Println("register listener:", id)
	wait := make(chan SyncMessage, 1)
	sync := make(chan struct{})
	c.acts <- func() {
		if _, ok := c.syncListeners[id]; !ok {
			c.syncListeners[id] = wait
		}
		sync <- struct{}{}
	}
	<-sync
}

func (c *MsgController) waitOnSyncListener(id ID) (msg SyncMessage, err error) {
	log.Println("wait on listener:", id)
	var (
		sync = make(chan struct{})
		wait chan SyncMessage
		ok   bool
	)
	c.acts <- func() {
		wait, ok = c.syncListeners[id]
		sync <- struct{}{}
	}
	<-sync
	if !ok {
		return nil, fmt.Errorf("unknown listener ID: %v", id)
	}
	select {
	case msg = <-wait:
	case <-time.After(c.ReplyTimeout):
		err = fmt.Errorf("timeout while waiting for message")
	}
	c.acts <- func() {
		delete(c.syncListeners, id)
	}
	return
}

func (c *MsgController) notifySyncListener(msg SyncMessage) {
	// pass in the request ID so we don't have to do any type assertion
	var (
		sync = make(chan struct{})
		l    chan SyncMessage
		ok   bool
	)
	c.acts <- func() {
		l, ok = c.syncListeners[msg.GetID()]
		sync <- struct{}{}
	}
	<-sync
	if ok {
		l <- msg
	} else {
		log.Println("no listener for message", msg.GetID())
	}
}
