package pub

import (
	"log"
	"xim/xchat/logic/pub/types"

	"github.com/go-mangos/mangos"
	"github.com/go-mangos/mangos/protocol/sub"
	"github.com/go-mangos/mangos/transport/ipc"
	"github.com/go-mangos/mangos/transport/tcp"
)

// Subscriber is a subscriber.
type Subscriber struct {
	s    mangos.Socket
	msgs chan interface{}
}

// NewSubscriber creates a Subscriber.
func NewSubscriber(addr string, bufSize int) *Subscriber {
	s, err := sub.NewSocket()
	if err != nil {
		log.Fatalf("new sub socket failed: %s\n", err)
	}
	s.AddTransport(ipc.NewTransport())
	s.AddTransport(tcp.NewTransport())

	if err := s.Dial(addr); err != nil {
		log.Fatal("can't dial on pub socket:", err)
	}
	log.Printf("sub dial to: %s\n", addr)
	s.SetOption(mangos.OptionSubscribe, []byte(""))

	sub := &Subscriber{
		s:    s,
		msgs: make(chan interface{}, bufSize),
	}

	go sub.subscribing()
	return sub
}

// Msgs is the message channel.
func (s *Subscriber) Msgs() <-chan interface{} {
	return s.msgs
}

func (s *Subscriber) subscribing() {
	for {
		buf, err := s.s.Recv()
		if err != nil {
			close(s.msgs)
			return
		}
		xmsg := &types.XMessage{}
		_, err = xmsg.Unmarshal(buf)
		if err != nil {
			// TODO:
			log.Println("decode message error:", err)
			continue
		}
		s.msgs <- xmsg.Msg
	}
}
