package jwtutils

import (
	"errors"
	"fmt"
	"strings"

	"github.com/dgrijalva/jwt-go"
)

// DecodeNSJwt decode ns/token from token string.
func DecodeNSJwt(rawToken string) (ns string, t string) {
	t = rawToken
	parts := strings.SplitN(rawToken, ":", 2)
	if len(parts) > 1 {
		ns = parts[0]
		t = parts[1]
	}

	token, _ := jwt.Parse(t, nil)
	claims, ok := token.Claims.(jwt.MapClaims)
	// 优先使用claims中的ns
	if ok {
		ns = claims["ns"].(string)
	}
	return ns, t
}

// ParseToken parse auth token with key to Token object.
func ParseToken(authToken string, key []byte) (jwt.MapClaims, error) {
	token, err := jwt.Parse(authToken, func(token *jwt.Token) (interface{}, error) {
		// Check the signing method
		if token.Method.Alg() != jwt.SigningMethodHS256.Name {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Method.Alg())
		}
		return key, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

// ParseNsToken parse ns auth token with keys to ns and Token object.
func ParseNsToken(nsToken string, keys map[string][]byte) (string, jwt.MapClaims, error) {
	ns, t := DecodeNSJwt(nsToken)
	key, ok := keys[ns]
	if !ok {
		return "", nil, errors.New("ns not exist")
	}

	claims, err := ParseToken(t, key)
	return ns, claims, err
}
