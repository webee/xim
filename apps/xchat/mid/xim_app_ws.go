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
	ws        *websocket.Conn
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
	for {
		header := http.Header{}
		token := c.getToken()
		header.Add("Authorization", "Bearer "+token)
		ws, _, err := websocket.DefaultDialer.Dial(c.url, header)
		if err != nil {
			log.Println("websocket dial:", err)
			time.Sleep(1 * time.Second)
			continue
		}
		defer ws.Close()
		c.ws = ws
		c.handleMsg()
		c.ws = nil
	}
}

// SendMsg send msg.
func (c *XIMAppWsConn) SendMsg(v interface{}) {
	// TODO use a buffer channel.
	c.ws.WriteJSON(v)
}

func (c *XIMAppWsConn) handleMsg() {
	for {
		msg := make(map[string]interface{})
		if err := c.ws.ReadJSON(&msg); err != nil {
			break
		}
	}
}
