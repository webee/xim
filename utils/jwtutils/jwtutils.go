package jwtutils

import (
	"fmt"

	"github.com/dgrijalva/jwt-go"
)

// ParseToken parse auth token with key to Token object.
func ParseToken(authToken string, key []byte) (*jwt.Token, error) {
	return jwt.Parse(authToken, func(token *jwt.Token) (interface{}, error) {
		// Check the signing method
		if token.Method.Alg() != jwt.SigningMethodHS256.Name {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Method.Alg())
		}
		return key, nil
	})
}
