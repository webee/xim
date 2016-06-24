package server

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
)

// RequireIsAdminUser ensure user is admin.
func RequireIsAdminUser(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := c.Get("token").(*jwt.Token)
		x := token.Claims.(jwt.MapClaims)["is_admin"]
		if x == nil {
			return echo.ErrUnauthorized
		}
		if x.(bool) {
			return next(c)
		}
		return echo.ErrUnauthorized
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
func JWT(contextKey string, key []byte) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			auth, err := jwtFromHeaderOrQueryParam(c)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, err.Error())
			}
			token, err := jwt.Parse(auth, func(t *jwt.Token) (interface{}, error) {
				// Check the signing method
				if t.Method.Alg() != "HS256" {
					return nil, fmt.Errorf("unexpected jwt signing method=%v", t.Header["alg"])
				}
				return key, nil

			})
			if err == nil && token.Valid {
				// Store user information from token into context.
				c.Set(contextKey, token)
				return next(c)
			}
			return echo.ErrUnauthorized
		}
	}
}
