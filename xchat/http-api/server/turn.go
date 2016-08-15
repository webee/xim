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

func fetchTurnServers(c echo.Context) error {
	user := config.TurnUser
	realm := "qqwj.com"
	secret := config.TurnSecret
	ttl := int64(3600)

	h := md5.New()
	io.WriteString(h, fmt.Sprintf("%s:%s:%s", user, realm, secret))
	key := h.Sum(nil)
	username := fmt.Sprintf("%d:%s", time.Now().Unix()+ttl, user)

	mac := hmac.New(sha1.New, key)
	io.WriteString(mac, username)
	password := mac.Sum(nil)

	servers := map[string]interface{}{
		"username": username,
		"password": password,
		"ttl":      ttl,
		"uris": []string{
			"turn:t.turn.engdd.com:3478?transport=udp",
			"turn:t.turn.engdd.com:3478?transport=tcp",
		},
	}
	return c.JSON(http.StatusOK, servers)
}
