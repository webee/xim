package httpapi

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
)

func fetchApp(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		appToken := c.Get("appToken").(*jwt.Token)
		app := appToken.Claims["app"].(string)
		c.Set("app", app)
		return next(c)
	}
}
