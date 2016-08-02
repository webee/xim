package turnpike

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// errors.
var (
	ErrWSSendTimeout = errors.New("ws peer send timeout")
	ErrWSIsClosed    = errors.New("ws peer is closed")
)

type websocketPeer struct {
	sync.Mutex
	conn        *websocket.Conn
	serializer  Serializer
	sendMsgs    chan Message
	messages    chan Message
	payloadType int
	inSending   chan struct{}
	closing     chan struct{}
	*ConnectionConfig
}

// NewWebsocketPeer connects to the websocket server at the specified url.
func NewWebsocketPeer(serialization Serialization, url, origin string, tlscfg *tls.Config) (Peer, error) {
	switch serialization {
	case JSON:
		return newWebsocketPeer(url, jsonWebsocketProtocol, origin,
			new(JSONSerializer), websocket.TextMessage, tlscfg,
		)
	case MSGPACK:
		return newWebsocketPeer(url, msgpackWebsocketProtocol, origin,
			new(MessagePackSerializer), websocket.BinaryMessage, tlscfg,
		)
	default:
		return nil, fmt.Errorf("Unsupported serialization: %v", serialization)
	}
}

func newWebsocketPeer(url, protocol, origin string, serializer Serializer, payloadType int, tlscfg *tls.Config) (Peer, error) {
	dialer := websocket.Dialer{
		Subprotocols:    []string{protocol},
		TLSClientConfig: tlscfg,
	}
	conn, _, err := dialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}
	ep := &websocketPeer{
		conn:             conn,
		sendMsgs:         make(chan Message, 16),
		messages:         make(chan Message, 10),
		serializer:       serializer,
		payloadType:      payloadType,
		closing:          make(chan struct{}),
		ConnectionConfig: &ConnectionConfig{},
	}
	go ep.run()

	return ep, nil
}

func (ep *websocketPeer) Send(msg Message) error {
	select {
	case ep.sendMsgs <- msg:
		return nil
	case <-time.After(5 * time.Second):
		log.Println(ErrWSSendTimeout.Error())
		// 发送太慢, 不正常的连接
		ep.Close()
		return ErrWSSendTimeout
	case <-ep.closing:
		log.Println(ErrWSIsClosed.Error())
		return ErrWSIsClosed
	}
}

func (ep *websocketPeer) Receive() <-chan Message {
	return ep.messages
}

func (ep *websocketPeer) doClosing() {
	select {
	case <-ep.closing:
	default:
		close(ep.closing)
	}
}

func (ep *websocketPeer) isClosed() bool {
	select {
	case <-ep.closing:
		return true
	default:
		return false
	}
}

func (ep *websocketPeer) Close() error {
	if ep.isClosed() {
		return nil
	}
	ep.doClosing()

	if ep.inSending != nil {
		<-ep.inSending
	}

	closeMsg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "goodbye")
	err := ep.conn.WriteControl(websocket.CloseMessage, closeMsg, time.Now().Add(5*time.Second))
	if err != nil {
		tlog.Println("error sending close message:", err)
	}

	return ep.conn.Close()
}

func (ep *websocketPeer) updateReadDeadline() {
	ep.Lock()
	defer ep.Unlock()
	if ep.IdleTimeout > 0 {
		ep.conn.SetReadDeadline(time.Now().Add(ep.IdleTimeout))
	}
}

func (ep *websocketPeer) setReadDead() {
	ep.Lock()
	defer ep.Unlock()
	ep.conn.SetReadDeadline(time.Now())
}

func (ep *websocketPeer) run() {
	go ep.sending()

	if ep.MaxMsgSize > 0 {
		ep.conn.SetReadLimit(ep.MaxMsgSize)
	}
	ep.conn.SetPongHandler(func(v string) error {
		tlog.Println("pong:", v)
		ep.updateReadDeadline()
		return nil
	})
	ep.conn.SetPingHandler(func(v string) error {
		tlog.Println("ping:", v)
		ep.updateReadDeadline()
		return nil
	})

	for {
		// TODO: use conn.NextMessage() and stream
		// TODO: do something different based on binary/text frames
		ep.updateReadDeadline()
		if msgType, b, err := ep.conn.ReadMessage(); err != nil {
			if ep.isClosed() {
				tlog.Println("peer connection closed")
			} else {
				tlog.Println("error reading from peer:", err)
			}
			close(ep.messages)
			break
		} else if msgType == websocket.CloseMessage {
			close(ep.messages)
			break
		} else {
			msg, err := ep.serializer.Deserialize(b)
			if err != nil {
				tlog.Println("error deserializing peer message:", err)
				// TODO: handle error
			} else {
				ep.messages <- msg
			}
		}
	}
}

func (ep *websocketPeer) sending() {
	ep.inSending = make(chan struct{})
	var ticker *time.Ticker
	if ep.PingTimeout == 0 {
		ticker = time.NewTicker(7 * 24 * time.Hour)
	} else {
		ticker = time.NewTicker(ep.PingTimeout)
	}

	defer func() {
		ep.setReadDead()
		ticker.Stop()
		close(ep.inSending)
	}()

	for {
		select {
		case msg := <-ep.sendMsgs:
			if closed, _ := ep.doSend(msg); closed {
				return
			}
		case <-ticker.C:
			if err := ep.conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(ep.WriteTimeout)); err != nil {
				return
			}
		case <-ep.closing:
			// sending remaining messages.
			for {
				select {
				case msg := <-ep.sendMsgs:
					if closed, _ := ep.doSend(msg); !closed {
						continue
					}
				default:
				}
				break
			}
			return
		}
		ep.updateReadDeadline()
	}
}

func (ep *websocketPeer) doSend(msg Message) (closed bool, err error) {
	tlog.Printf("do sending message: %s< %+v >", msg.MessageType(), msg)
	b, err := ep.serializer.Serialize(msg)
	if err != nil {
		log.Printf("error serializing peer message: %s, %+v", err, msg)
		return true, err
	}
	if ep.WriteTimeout > 0 {
		ep.conn.SetWriteDeadline(time.Now().Add(ep.WriteTimeout))
	}
	if err = ep.conn.WriteMessage(ep.payloadType, b); err != nil {
		tlog.Println("error write message: ", err)
		return true, err
	}
	return
}
