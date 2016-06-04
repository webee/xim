package nanorpc

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"math/rand"
	"net/rpc"
	"sync"
	"time"

	"github.com/go-mangos/mangos"
)

const (
	maxID int64 = 1 << 53
)

// NewID generates a random request ID.
func NewID() uint64 {
	return uint64(rand.Int63n(maxID))
}

type req struct {
	msg *mangos.Message
	id  uint32
}

type nanoGobServerCodec struct {
	sync.Mutex
	s      mangos.Socket
	reqs   map[uint64]*req
	dec    *gob.Decoder
	closed bool
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
	if err := s.SetOption(mangos.OptionSendDeadline, 3*time.Second); err != nil {
		log.Panic(err)
	}
	return &nanoGobServerCodec{
		s:    s,
		reqs: make(map[uint64]*req),
	}
}

func (c *nanoGobServerCodec) ReadRequestHeader(r *rpc.Request) error {
	msg, err := c.s.RecvMsg()
	// FIXME: this must be a pipe.
	if err != nil {
		return err
	}
	dec := gob.NewDecoder(bytes.NewBuffer(msg.Body))
	err = dec.Decode(r)
	if err != nil {
		r.Seq = NewID()
	} else {
		c.dec = dec
	}
	id := interface{}(msg.Port).(pipe).GetID()
	r.Seq = r.Seq ^ (uint64(id) << 32)

	c.Lock()
	c.reqs[r.Seq] = &req{msg, id}
	c.Unlock()

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
	c.Lock()
	req, ok := c.reqs[r.Seq]
	delete(c.reqs, r.Seq)
	c.Unlock()
	if !ok {
		return fmt.Errorf("request missing")
	}

	r.Seq = r.Seq ^ (uint64(req.id) << 32)

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
	c.closed = true
	return c.s.Close()
}

type nanoGobClientCodec struct {
	sync.Mutex
	s      mangos.Socket
	dec    *gob.Decoder
	closed bool
	nextid uint32
}

func (c *nanoGobClientCodec) nextID() uint32 {
	c.Lock()
	defer c.Unlock()
	// The high order bit is "special", and must always be set.  (This is
	// how the peer will detect the end of the backtrace.)
	v := c.nextid | 0x80000000
	c.nextid++
	return v
}

// NewNanoGobClientCodec returns a new rpc.ClientCodec.
func NewNanoGobClientCodec(s mangos.Socket) rpc.ClientCodec {
	if err := s.SetOption(mangos.OptionSendDeadline, 3*time.Second); err != nil {
		log.Panic(err)
	}
	if err := s.SetOption(mangos.OptionRecvDeadline, 5*time.Second); err != nil {
		log.Panic(err)
	}
	return &nanoGobClientCodec{
		s:      s,
		nextid: uint32(time.Now().UnixNano()),
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
	return c.s.Close()
}
