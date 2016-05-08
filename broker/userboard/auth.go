package userboard

import (
	"fmt"

	"xim/broker/userds"

	"github.com/dgrijalva/jwt-go"
)

// VerifyAppAuthToken verify app token.
func VerifyAppAuthToken(auth string) (aid *userds.AppIdentity, err error) {
	token, err := parseToken(auth, appKey)

	if err != nil || !token.Valid {
		return nil, err
	}

	aid = &userds.AppIdentity{
		App: token.Claims["app"].(string),
	}
	return aid, nil
}

// VerifyAuthToken verify user token.
func VerifyAuthToken(auth string) (uid *userds.UserIdentity, err error) {
	token, err := parseToken(auth, userKey)
	if err != nil || !token.Valid {
		return nil, err
	}
	uid = &userds.UserIdentity{
		App:  token.Claims["app"].(string),
		User: token.Claims["user"].(string),
	}
	return uid, nil
}

func parseToken(authToken string, key []byte) (*jwt.Token, error) {
	return jwt.Parse(authToken, func(token *jwt.Token) (interface{}, error) {
		// Check the signing method
		if token.Method.Alg() != jwt.SigningMethodHS256.Name {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Method.Alg())
		}
		return key, nil
	})
}
