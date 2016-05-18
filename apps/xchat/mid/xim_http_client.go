package mid

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// api uris.
const (
	URIAppNewToken = "/xim/app.new_token"
)

var (
	ximHTTPClient *XIMHTTPClient
)

func initXimHTTPClient(app, password, hostURL string) {
	ximHTTPClient = NewXIMHTTPClient(app, password, hostURL)
}

// XIMHTTPClient is the xim http api client.
type XIMHTTPClient struct {
	sync.Mutex
	app      string
	password string
	hostURL  string
	token    string
	tokenExp int64
}

// NewXIMHTTPClient creates a xim client.
func NewXIMHTTPClient(app, password string, hostURL string) *XIMHTTPClient {
	return &XIMHTTPClient{
		app:      app,
		password: password,
		hostURL:  hostURL,
	}
}

func (c *XIMHTTPClient) url(uri string) string {
	return c.hostURL + uri
}

// NewToken request a new app token.
func (c *XIMHTTPClient) NewToken() (string, error) {
	resp, err := http.PostForm(c.url(URIAppNewToken), url.Values{
		"username": {c.app},
		"password": {c.password},
	})
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", errors.New("request failed")
	}
	decoder := json.NewDecoder(resp.Body)
	res := make(map[string]string)
	if err := decoder.Decode(&res); err != nil {
		return "", err
	}
	return res["token"], nil
}

func (c *XIMHTTPClient) isCurrentTokenValid() bool {
	n := time.Now().Unix()
	return c.token != "" && c.tokenExp > n
}

// Token returns a valid token.
func (c *XIMHTTPClient) Token() string {
	if !c.isCurrentTokenValid() {
		c.Lock()
		defer c.Unlock()
		if !c.isCurrentTokenValid() {
			token, err := c.NewToken()
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
