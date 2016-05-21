package mid

import (
	"errors"
	"log"
	"sync"
	"time"
	"xim/broker/proto"
	"xim/utils/msgutils"
)

// XIMClient handles xim apis and app ws connections.
type XIMClient struct {
	sync.RWMutex
	config *Config
	// sessionID to uid.
	sessionUids        map[uint64]uint32
	uidSessions        map[uint32]uint64
	handler            msgutils.MessageHandler
	ximAppWsController *XIMAppWsController
	done               chan struct{}
}

// NewXIMClient create a xim client.
func NewXIMClient(config *Config, handler msgutils.MessageHandler) *XIMClient {
	x := &XIMClient{
		config:      config,
		sessionUids: make(map[uint64]uint32, 1024),
		uidSessions: make(map[uint32]uint64, 1024),
		handler:     handler,
		done:        make(chan struct{}, 1),
	}
	x.ximAppWsController = x.newXimAppWsController()
	x.ximAppWsController.Start()
	go x.ping()
	return x
}

// Register register user with sessionID.
func (c *XIMClient) Register(sessionID uint64, user string) error {
	r, err := c.ximAppWsController.Req(&proto.Register{
		User: user,
	})
	if err != nil {
		return err
	}

	c.Lock()
	defer c.Unlock()

	reply := r.(*proto.Reply)
	if !reply.Ok {
		return errors.New("register failed")
	}
	c.sessionUids[sessionID] = reply.UID
	c.uidSessions[reply.UID] = sessionID
	return nil
}

// Unregister unregister sessionID user.
func (c *XIMClient) Unregister(sessionID uint64) error {
	uid, err := c.getUIDbySessionID(sessionID)
	if err != nil {
		return err
	}

	r, err := c.ximAppWsController.Req(&proto.Unregister{
		UID: uid,
	})
	if err != nil {
		return err
	}
	reply := r.(*proto.Reply)
	if !reply.Ok {
		return errors.New("unregister failed")
	}
	delete(c.uidSessions, uid)
	delete(c.sessionUids, sessionID)
	return nil
}

// SendMsg send uid's msg to channel.
func (c *XIMClient) SendMsg(sessionID uint64, channel string, msg interface{}) (id uint64, ts uint64, err error) {
	uid, err := c.getUIDbySessionID(sessionID)
	if err != nil {
		return
	}

	r, err := c.ximAppWsController.Req(&proto.Put{
		UID:     uid,
		Channel: channel,
		Msg:     msg,
	})
	if err != nil {
		return
	}

	reply := r.(*proto.Reply)
	if !reply.Ok {
		err = errors.New("send msg failed")
		return
	}
	data := reply.Data.(map[string]interface{})
	id = uint64(data["id"].(float64))
	ts = uint64(data["ts"].(float64))
	return
}

// Close free resources.
func (c *XIMClient) Close() {
	c.ximAppWsController.Close()
	close(c.done)
}

func (c *XIMClient) getSessionIDbyUID(uid uint32) (uint64, error) {
	c.RLock()
	defer c.RUnlock()
	sessionID, ok := c.uidSessions[uid]
	if !ok {
		return 0, errors.New("user session not found")
	}
	return sessionID, nil
}

func (c *XIMClient) getUIDbySessionID(sessionID uint64) (uint32, error) {
	c.RLock()
	defer c.RUnlock()
	uid, ok := c.sessionUids[sessionID]
	if !ok {
		return 0, errors.New("session user not found")
	}
	return uid, nil
}

func (c *XIMClient) ping() {
	ticker := time.NewTicker(32 * time.Second)
	for t := range ticker.C {
		select {
		case <-c.done:
			return
		default:
			log.Println("ping at: ", t)
			_ = c.ximAppWsController.Send(proto.PING.New())
		}
	}
}

func (c *XIMClient) newXimAppWsController() *XIMAppWsController {
	for {
		token := ximHTTPClient.Token()
		if t, err := getWSTranseiver(c.config.XIMAppWsURL, token, 1024); err == nil {
			return NewXIMAppWsController(t, c.handler)
		}
		time.Sleep(2 * time.Second)
	}
}
