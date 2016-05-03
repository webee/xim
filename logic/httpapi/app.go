package httpapi

import (
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
)

// newUserToken is an app api to generate a user access token.
func newUserToken(c echo.Context) (err error) {
	appToken := c.Get("app").(*jwt.Token)
	app := appToken.Claims["app"].(string)
	user := c.QueryParam("user")
	expire := c.QueryParam("expire")

	if user == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"ok":  false,
			"err": "bad user",
		})
	}
	exp := 10 * time.Minute
	if expire != "" {
		exp, err = time.ParseDuration(expire)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"ok":  false,
				"err": "bad expire time",
			})
		}
	}

	// Create token
	token := jwt.New(jwt.SigningMethodHS256)
	expireAt := time.Now().Add(exp).Unix()

	// Set claims
	token.Claims["app"] = app
	token.Claims["user"] = user
	token.Claims["exp"] = expireAt

	// Generate encoded token and send it as response.
	t, err := token.SignedString(userKey)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"ok":        true,
		"expire_at": expireAt,
		"token":     t,
	})
}

// newChannel creates a messaging channel.
func newChannel(c echo.Context) error {
	return c.String(http.StatusOK, "TODO")
}

// channelAddMembers add members to channel.
func channelAddMembers(c echo.Context) error {
	return c.String(http.StatusOK, "TODO")
}
