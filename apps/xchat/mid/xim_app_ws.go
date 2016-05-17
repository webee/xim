package mid

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/websocket"
)

var (
	ximHTTPClient *XIMHTTPClient
)

func ximInitSetup(config *Config) {
	ximAppWsConn := NewXIMAppWsConn(config.XIMApp, config.XIMPassword, config.XIMHostURL, config.XIMAppWsURL)
	ximAppWsConn.run()
}

// XIMAppWsConn is a app websocket connection to xim.
type XIMAppWsConn struct {
	sync.Mutex
	ximClient *XIMHTTPClient
	url       string
	token     string
	tokenExp  int64
	conn      Connection
}

// NewXIMAppWsConn creates a xim app websocket connection object.
func NewXIMAppWsConn(app, password, hostURL, wsURL string) *XIMAppWsConn {
	conn := &XIMAppWsConn{
		ximClient: NewXIMHTTPClient(app, password, hostURL),
		url:       wsURL,
	}
	go conn.run()
	return conn
}

func (c *XIMAppWsConn) isCurrentTokenValid() bool {
	n := time.Now().Unix()
	return c.token != "" && c.tokenExp > n
}

func (c *XIMAppWsConn) getToken() string {
	if !c.isCurrentTokenValid() {
		c.Lock()
		defer c.Unlock()
		if !c.isCurrentTokenValid() {
			token, err := c.ximClient.NewToken()
			if err != nil {
				log.Println("get token:", err)
				return ""
			}
			c.token = token
			t, _ := jwt.Parse(token, nil)
			c.tokenExp = int64(t.Claims["exp"].(float64))
		}
	}
	return c.token
}

func (c *XIMAppWsConn) run() {
	var retryTimes time.Duration
	for retryTimes <= 3 {
		header := http.Header{}
		token := c.getToken()
		header.Add("Authorization", "Bearer "+token)
		wsConn, _, err := websocket.DefaultDialer.Dial(c.url, header)
		if err != nil {
			log.Println("websocket dial:", err)
			time.Sleep(2 * retryTimes * time.Second)
			retryTimes++
			continue
		}
		retryTimes = 0

		conn := newWsConnection(wsConn, 100)
		c.conn = conn
		c.handleMsg()
		c.conn = nil
	}
}

// SendMsg send msg.
func (c *XIMAppWsConn) SendMsg(v interface{}) {
	// TODO use a buffer channel.
	c.conn.Send(v)
}

func (c *XIMAppWsConn) handleMsg() {
	defer c.conn.Close()
	var msg map[string]interface{}

	msg, err := GetMessageTimeout(c.conn, 2*time.Second)
	if err != nil {
		log.Println("waiting hello error:", err)
	}
	msgType, ok := msg["type"].(string)
	if !ok || msgType != "hello" {
		log.Println("not hello msg.")
	}

	r := c.conn.Receive()
	for {
		var msg map[string]interface{}
		var open bool
		msg, open = <-r
		if !open {
			return
		}
		log.Println("msg:", msg)
	}
}
