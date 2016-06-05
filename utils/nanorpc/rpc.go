package nanorpc

import (
	"errors"
	"log"
	"net/rpc"
	"sync"
	"time"

	"github.com/go-mangos/mangos"
	"github.com/go-mangos/mangos/protocol/rep"
	"github.com/go-mangos/mangos/protocol/req"
	"github.com/go-mangos/mangos/transport/ipc"
	"github.com/go-mangos/mangos/transport/tcp"
)

// ErrRPCTimeout represents waiting rpc call result timeout.
var ErrRPCTimeout = errors.New("rpc call timeout")

// RPCCallTimeout is the rpc call timeout.
var RPCCallTimeout = 5 * time.Second

// RPCSendTimeout is the rpc send timeout.
var RPCSendTimeout = 3 * time.Second

// Client is a rpc client.
type Client struct {
	sync.Mutex
	*rpc.Client
	addr          string
	connectedTime time.Time
}

// Call call a method.
func (r *Client) Call(serviceMethod string, args interface{}, reply interface{}) error {
	err := r.CallTimeout(serviceMethod, RPCCallTimeout, args, reply)
	if err == rpc.ErrShutdown {
		r.reconnect()
		return r.CallTimeout(serviceMethod, RPCCallTimeout, args, reply)
	}
	return err
}

// CallTimeout call a method within timeout.
func (r *Client) CallTimeout(serviceMethod string, t time.Duration, args interface{}, reply interface{}) error {
	select {
	case call := <-r.Client.Go(serviceMethod, args, reply, make(chan *rpc.Call, 1)).Done:
		return call.Error
	case <-time.After(t):
		return ErrRPCTimeout
	}
}

func (r *Client) reconnect() {
	r.Lock()
	defer r.Unlock()
	if time.Now().Sub(r.connectedTime) < 2*time.Second {
		return
	}

	r.Client.Close()
	r.Client = newGoRPCClient(r.addr)
	r.connectedTime = time.Now()
}

// NewClient return s rpc client dial to addr.
func NewClient(addr string) *Client {
	return &Client{
		Client:        newGoRPCClient(addr),
		addr:          addr,
		connectedTime: time.Now(),
	}
}

func newGoRPCClient(addr string) *rpc.Client {
	s, err := req.NewSocket()
	if err != nil {
		log.Fatal("failed to open socket:", err)
	}

	s.SetOption(mangos.OptionRaw, true)
	s.AddTransport(tcp.NewTransport())
	s.AddTransport(ipc.NewTransport())
	if err := s.Dial(addr); err != nil {
		log.Fatal("can't dial on socket:", err)
	}
	log.Printf("rpc dial to: %s\n", addr)

	if err := s.SetOption(mangos.OptionSendDeadline, RPCSendTimeout); err != nil {
		log.Panic(err)
	}

	return rpc.NewClientWithCodec(NewNanoGobClientCodec(s))
}

// StartRPCServer starts rpc server with register rcvrs.
func StartRPCServer(addrs []string, dial bool, rcvrs ...interface{}) (close func()) {
	for _, rcvr := range rcvrs {
		rpc.Register(rcvr)
	}
	s := getReplySocket(addrs, dial)
	codec := NewNanoGobServerCodec(s)
	go rpc.ServeCodec(codec)
	return func() {
		codec.Close()
	}
}

func getReplySocket(addrs []string, dial bool) mangos.Socket {
	s, err := rep.NewSocket()
	if err != nil {
		log.Fatal("failed to open reply socket:", err)
	}

	s.SetOption(mangos.OptionRaw, true)
	s.AddTransport(tcp.NewTransport())
	s.AddTransport(ipc.NewTransport())

	if dial {
		// dial to load balancing rep/req proxy.
		for _, addr := range addrs {
			if err := s.Dial(addr); err != nil {
				log.Fatal("can't dial on request socket:", err)
			}
			log.Printf("rpc dial to: %s\n", addr)

		}
	} else {
		// serve
		for _, addr := range addrs {
			if err := s.Listen(addr); err != nil {
				log.Fatal("can't listen on reply socket:", err)
			}
			log.Printf("rpc listen on: %s\n", addr)
		}
	}
	if err := s.SetOption(mangos.OptionSendDeadline, RPCSendTimeout); err != nil {
		log.Panic(err)
	}
	return s
}
