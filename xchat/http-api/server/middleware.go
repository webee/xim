package server

import (
	"errors"
	"net/http"
	"xim/utils/jwtutils"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
)

// GetContextString get string value from echo context.
func GetContextString(key string, c echo.Context) string {
	return c.Get(key).(string)
}

// RequireIsAdminUser ensure user is admin.
func RequireIsAdminUser(tokenKey string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			claims := c.Get(tokenKey).(jwt.MapClaims)
			x := claims["is_admin"]
			xb, ok := x.(bool)
			if !ok || !xb {
				return echo.ErrUnauthorized
			}
			return next(c)
		}
	}
}

func jwtFromHeader(c echo.Context) (string, error) {
	auth := c.Request().Header().Get("Authorization")
	l := len("Bearer")
	if len(auth) > l+1 && auth[:l] == "Bearer" {
		return auth[l+1:], nil
	}
	return "", errors.New("empty or invalid jwt in authorization header")
}

func jwtFromHeaderOrQueryParam(c echo.Context) (string, error) {
	token, err := jwtFromHeader(c)
	if err != nil {
		token = c.FormValue("jwt")
	}
	if token != "" {
		return token, nil
	}
	return "", err
}

// JWT checks jwt.
func JWT(nsKey string, tokenKey string, keys map[string][]byte) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			auth, err := jwtFromHeaderOrQueryParam(c)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, err.Error())
			}
			ns, claims, err := jwtutils.ParseNsToken(auth, keys)
			if err != nil {
				return echo.ErrUnauthorized
			}
			// Store user information from token into context.
			c.Set(nsKey, ns)
			c.Set(tokenKey, claims)
			return next(c)
		}
	}
}
