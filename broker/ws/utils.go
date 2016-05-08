package ws

import (
	"errors"
	"net/http"
	"strings"
)

func getAuthTokenFromRequest(r *http.Request) (token string, err error) {
	bearerAuth := r.Header.Get("Authorization")
	if bearerAuth != "" {
		parts := strings.SplitN(bearerAuth, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			return "", errors.New("invalid jwt authorization header=" + bearerAuth)
		}
		token = parts[1]
	} else {
		token = r.FormValue("jwt")
	}
	if token == "" {
		err = errors.New("need authorization token")
	}
	return
}
