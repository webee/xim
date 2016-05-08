package httpapi

import (
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
)

// putMsg put message to channel.
func putMsg(c echo.Context) error {
	appToken := c.Get("app").(*jwt.Token)
	app := appToken.Claims["app"].(string)
	return c.String(http.StatusOK, "APP:"+app)
}
