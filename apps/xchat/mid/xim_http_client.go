package mid

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
)

// api uris.
const (
	URIAppNewToken = "/xim/app.new_token"
)

// XIMHTTPClient is the xim http api client.
type XIMHTTPClient struct {
	app      string
	password string
	hostURL  string
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
