package mid

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// api uris.
const (
	URIAppNewToken      = "/xim/app.new_token"
	URIFetchChannelMsgs = "/xim/app/channels/%s.msgs?%s"
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
	client   *http.Client
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
		client:   &http.Client{},
		password: password,
		hostURL:  hostURL,
	}
}

func (c *XIMHTTPClient) url(uri string, params ...interface{}) string {
	return c.hostURL + fmt.Sprintf(uri, params...)
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

// FetchChannelMsgs fetch channel's messages.
func (c *XIMHTTPClient) FetchChannelMsgs(channel string, lid, rid uint64, limit int) ([]UserMsg, error) {
	params := url.Values{}
	if lid > 0 {
		params.Add("lid", strconv.FormatUint(lid, 10))
	}
	if rid > 0 {
		params.Add("rid", strconv.FormatUint(rid, 10))
	}
	if limit > 0 {
		params.Add("rid", strconv.Itoa(limit))
	}

	req, err := http.NewRequest("GET", c.url(URIFetchChannelMsgs, channel, params.Encode()), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+c.Token())

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, errors.New("request failed")
	}
	decoder := json.NewDecoder(resp.Body)
	res := []UserMsg{}
	if err := decoder.Decode(&res); err != nil {
		return nil, err
	}
	return res, nil
}
