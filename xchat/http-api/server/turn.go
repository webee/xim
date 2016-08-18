package server

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/labstack/echo"
)

// IceConfig is ice config.
type IceConfig struct {
	IceServers []map[string]interface{} `json:"iceServers"`
	TTL        int64                    `json:"ttl"`
}

func fetchIceConfig(c echo.Context) error {
	// TODO: add user/secret params, add token check.
	// TODO: check ip area, return the best turn servers.
	user := config.TurnUser
	realm := "qqwj.com"
	secret := config.TurnSecret
	ttl := config.TurnPasswordTTL

	h := md5.New()
	io.WriteString(h, fmt.Sprintf("%s:%s:%s", user, realm, secret))
	key := h.Sum(nil)
	username := fmt.Sprintf("%d:%s", time.Now().Unix()+ttl, user)

	mac := hmac.New(sha1.New, key)
	io.WriteString(mac, username)
	credential := mac.Sum(nil)

	iceConfig := IceConfig{
		TTL: ttl,
		IceServers: []map[string]interface{}{
			map[string]interface{}{
				"urls": []string{
					fmt.Sprintf("stun:%s", config.TurnURI),
				},
			},
			map[string]interface{}{
				"username":   username,
				"credential": credential,
				"urls": []string{
					fmt.Sprintf("turn:%s?transport=udp", config.TurnURI),
					fmt.Sprintf("turn:%s?transport=tcp", config.TurnURI),
				},
			},
		},
	}
	return c.JSON(http.StatusOK, iceConfig)
}
