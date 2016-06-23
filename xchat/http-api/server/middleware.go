package server

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
)

// RequireIsAdminUser ensure user is admin.
func RequireIsAdminUser(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := c.Get("token").(*jwt.Token)
		x := token.Claims["is_admin"]
		if x == nil {
			return echo.ErrUnauthorized
		}
		if x.(bool) {
			return next(c)
		}
		return echo.ErrUnauthorized
	}
}
