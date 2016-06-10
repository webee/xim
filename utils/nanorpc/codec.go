package nanorpc

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"math"
	"math/rand"
	"net/rpc"
	"time"

	"github.com/go-mangos/mangos"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// NewID generates a random request ID.
func NewID() uint64 {
	return uint64(rand.Int63n(math.MaxInt64))
}

type request struct {
	msg *mangos.Message
	seq uint64
}

type nanoGobServerCodec struct {
	s      mangos.Socket
	reqs   map[uint64]*request
	dec    *gob.Decoder
	closed bool
	acts   chan func()
}

type pipe interface {
	GetID() uint32
}

// NewNanoGobServerCodec returns a new rpc.ServerCodec.
//
// A ServerCodec implements reading of RPC requests and writing of RPC
// responses for the server side of an RPC session. The server calls
// ReadRequestHeader and ReadRequestBody in pairs to read requests from the
// connection, and it calls WriteResponse to write a response back. The
// server calls Close when finished with the connection.
func NewNanoGobServerCodec(s mangos.Socket) rpc.ServerCodec {
	c := &nanoGobServerCodec{
		s:    s,
		reqs: make(map[uint64]*request),
		acts: make(chan func()),
	}
	go c.run()
	return c
}

func (c *nanoGobServerCodec) run() {
	for act := range c.acts {
		act()
	}
}

func (c *nanoGobServerCodec) ReadRequestHeader(r *rpc.Request) error {
	id := NewID()
	msg, err := c.s.RecvMsg()
	if err != nil {
		return err
	}
	dec := gob.NewDecoder(bytes.NewBuffer(msg.Body))
	err = dec.Decode(r)
	if err != nil {
		return err
	}
	c.dec = dec

	//log.Printf("id: %d, seq: %d, %+v\n", id, r.Seq, msg.Header)
	sync := make(chan struct{})
	c.acts <- func() {
		c.reqs[id] = &request{msg, r.Seq}
		sync <- struct{}{}
	}
	<-sync
	r.Seq = id

	return err
}

func (c *nanoGobServerCodec) ReadRequestBody(body interface{}) error {
	if c.dec == nil {
		return fmt.Errorf("no decoder")
	}
	dec := c.dec
	c.dec = nil
	return dec.Decode(body)
}

func (c *nanoGobServerCodec) WriteResponse(r *rpc.Response, body interface{}) (err error) {
	var (
		req *request
		ok  bool
	)
	sync := make(chan struct{})
	c.acts <- func() {
		req, ok = c.reqs[r.Seq]
		delete(c.reqs, r.Seq)
		sync <- struct{}{}
	}
	<-sync

	if !ok {
		return fmt.Errorf("request missing")
	}

	r.Seq = req.seq

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err = enc.Encode(r); err != nil {
		return
	}
	if err = enc.Encode(body); err != nil {
		return
	}
	req.msg.Body = buf.Bytes()
	return c.s.SendMsg(req.msg)
}

func (c *nanoGobServerCodec) Close() error {
	if c.closed {
		// Only call c.rwc.Close once; otherwise the semantics are undefined.
		return nil
	}
	close(c.acts)
	c.closed = true
	return c.s.Close()
}

type nanoGobClientCodec struct {
	s      mangos.Socket
	dec    *gob.Decoder
	closed bool
	nextid uint32
	acts   chan func()
}

func (c *nanoGobClientCodec) nextID() uint32 {
	var v uint32
	sync := make(chan struct{})
	c.acts <- func() {
		// The high order bit is "special", and must always be set.  (This is
		// how the peer will detect the end of the backtrace.)
		v = c.nextid | 0x80000000
		c.nextid++
		sync <- struct{}{}
	}
	<-sync
	return v
}

// NewNanoGobClientCodec returns a new rpc.ClientCodec.
func NewNanoGobClientCodec(s mangos.Socket) rpc.ClientCodec {
	c := &nanoGobClientCodec{
		s:      s,
		nextid: uint32(time.Now().UnixNano()),
		acts:   make(chan func()),
	}
	go c.run()
	return c
}

func (c *nanoGobClientCodec) run() {
	for act := range c.acts {
		act()
	}
}

func (c *nanoGobClientCodec) WriteRequest(r *rpc.Request, body interface{}) (err error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err = enc.Encode(r); err != nil {
		return
	}
	if err = enc.Encode(body); err != nil {
		return
	}
	msg := mangos.NewMessage(buf.Len())
	v := c.nextID()
	msg.Header = append(msg.Header, byte(v>>24), byte(v>>16), byte(v>>8), byte(v))
	msg.Body = append(msg.Body, buf.Bytes()...)

	//log.Printf("seq: %d, %+v\n", r.Seq, msg.Header)
	return c.s.SendMsg(msg)
}

func (c *nanoGobClientCodec) ReadResponseHeader(r *rpc.Response) error {
	body, err := c.s.Recv()
	if err != nil {
		return err
	}
	dec := gob.NewDecoder(bytes.NewBuffer(body))
	err = dec.Decode(r)
	if err != nil {
		return err
	}
	c.dec = dec
	return nil
}

func (c *nanoGobClientCodec) ReadResponseBody(body interface{}) error {
	if c.dec == nil {
		return fmt.Errorf("no decoder")
	}
	dec := c.dec
	c.dec = nil
	return dec.Decode(body)
}

func (c *nanoGobClientCodec) Close() error {
	if c.closed {
		return nil
	}
	close(c.acts)
	c.closed = true
	return c.s.Close()
}
