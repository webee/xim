package jwtutils

import (
	"errors"
	"fmt"

	"github.com/dgrijalva/jwt-go"
)

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
