package xchathttpclient

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// api uris.
const (
	URINewChat      = "/api/chats/"
	URIUserChatList = "/api/user/chats/?%s"
)

// XChatHTTPClient is the xim http api client.
// TODO: use multi  host urls to do load balance.
type XChatHTTPClient struct {
	sync.Mutex
	client   *http.Client
	userKey  []byte
	hostURL  string
	token    string
	tokenExp int64
}

// NewXChatHTTPClient creates a xim client.
func NewXChatHTTPClient(userKey []byte, hostURL string) *XChatHTTPClient {
	return &XChatHTTPClient{
		userKey: userKey,
		client:  &http.Client{},
		hostURL: hostURL,
	}
}

func (c *XChatHTTPClient) url(uri string, params ...interface{}) string {
	return c.hostURL + fmt.Sprintf(uri, params...)
}

// NewToken request a new app token.
func (c *XChatHTTPClient) newToken() (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"ns":       "",
		"is_admin": true,
		"exp":      time.Now().Add(30 * 24 * time.Hour).Unix(),
	})
	return token.SignedString(c.userKey)
}

func (c *XChatHTTPClient) isCurrentTokenValid() bool {
	n := time.Now().Unix()
	return c.token != "" && c.tokenExp > n
}

// Token returns a valid token.
func (c *XChatHTTPClient) getToken() string {
	if !c.isCurrentTokenValid() {
		c.Lock()
		defer c.Unlock()
		if !c.isCurrentTokenValid() {
			token, err := c.newToken()
			if err != nil {
				log.Println("get token:", err)
				return ""
			}
			c.token = token
			t, _ := jwt.Parse(token, nil)
			claims := t.Claims.(jwt.MapClaims)
			c.tokenExp = int64(claims["exp"].(float64))
		}
	}
	return c.token
}

// NewChat creates chat.
func (c *XChatHTTPClient) NewChat(chatType string, users []string, title, tag, ext string) (string, error) {
	params := make(map[string]interface{})
	params["type"] = chatType
	params["users"] = users
	params["tag"] = tag
	params["title"] = title
	params["ext"] = ext

	b, _ := json.Marshal(params)
	req, err := http.NewRequest("POST", c.url(URINewChat), bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	// NOTE: 利用了ns=""的情况可以添加任何ns的用户
	req.Header.Add("Authorization", "Bearer "+c.getToken())
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", errors.New("request failed")
	}
	decoder := json.NewDecoder(resp.Body)
	res := make(map[string]interface{})
	if err := decoder.Decode(&res); err != nil {
		return "", err
	}
	id, ok := res["id"]
	if !ok {
		return "", errors.New("request failed")
	}
	return id.(string), nil
}

// FetchUserChats fetch user's chat list.
func (c *XChatHTTPClient) FetchUserChats(user string, chatType string, tag string) ([]uint64, error) {
	params := url.Values{}
	params.Add("user", user)
	params.Add("type", chatType)
	params.Add("tag", tag)

	req, err := http.NewRequest("GET", c.url(URIUserChatList, params.Encode()), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+c.getToken())

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, errors.New("request failed")
	}
	decoder := json.NewDecoder(resp.Body)
	res := []struct {
		ID uint64 `json:"id"`
	}{}
	if err := decoder.Decode(&res); err != nil {
		return nil, err
	}

	ids := []uint64{}
	for _, r := range res {
		ids = append(ids, r.ID)
	}
	return ids, nil
}
