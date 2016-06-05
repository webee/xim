package pub

import (
	"log"
	"xim/xchat/logic/pub/types"

	"github.com/go-mangos/mangos"
	"github.com/go-mangos/mangos/protocol/pub"
	"github.com/go-mangos/mangos/transport/ipc"
	"github.com/go-mangos/mangos/transport/tcp"
)

var (
	s mangos.Socket
)

// StartPublisher starts the publisher.
func StartPublisher(addrs []string, dial bool) (close func()) {
	var err error
	s, err = pub.NewSocket()
	if err != nil {
		log.Fatal("failed to open reply socket:", err)
	}

	s.AddTransport(tcp.NewTransport())
	s.AddTransport(ipc.NewTransport())

	if dial {
		for _, addr := range addrs {
			if err := s.Dial(addr); err != nil {
				log.Fatal("can't dial on sub socket:", err)
			}
			log.Printf("pub dial to: %s\n", addr)

		}
	} else {
		for _, addr := range addrs {
			if err := s.Listen(addr); err != nil {
				log.Fatal("can't listen on pub socket:", err)
			}
			log.Printf("pub listen on: %s\n", addr)
		}
	}
	return func() {
		s.Close()
	}
}

// PublishMessage publish user send message.
func PublishMessage(msg *types.Message) error {
	b, err := msg.Marshal(nil)
	if err != nil {
		return err
	}

	return s.Send(b)
}
