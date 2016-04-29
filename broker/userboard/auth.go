package userboard

import (
	"fmt"

	"xim/broker/userds"

	"github.com/dgrijalva/jwt-go"
)

// VerifyAuthToken verify user token.
func VerifyAuthToken(auth string) (uid *userds.UserIdentity, err error) {
	token, err := jwt.Parse(auth, func(token *jwt.Token) (interface{}, error) {
		// Check the signing method
		if token.Method.Alg() != "HS256" {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Method.Alg())
		}
		return userKey, nil

	})
	if err != nil || !token.Valid {
		return nil, err
	}
	uid = &userds.UserIdentity{
		App:  token.Claims["app"].(string),
		User: token.Claims["user"].(string),
	}
	return uid, nil
}
