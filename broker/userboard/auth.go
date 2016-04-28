package userboard

import (
	"fmt"
	"strings"

	"github.com/dgrijalva/jwt-go"
)

// UserIdentity is a user instance.
type UserIdentity struct {
	App  string
	User string
}

func (uid UserIdentity) String() string {
	return fmt.Sprintf("%s:%s", uid.App, uid.User)
}

// ParseUserIdentify parse a user identity from a string.
func ParseUserIdentify(s string) *UserIdentity {
	parts := strings.Split(s, ":")
	return &UserIdentity{
		App:  parts[0],
		User: parts[1],
	}
}

// VerifyAuthToken verify user token.
func VerifyAuthToken(auth string) (uid *UserIdentity, err error) {
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
	uid = &UserIdentity{
		App:  token.Claims["app"].(string),
		User: token.Claims["user"].(string),
	}
	return uid, nil
}

// IsValid checks if user identity is valid.
func (uid *UserIdentity) IsValid() bool {
	return uid != nil && uid.App != "" && uid.User != ""
}

// ResetTimeout reset user identity timeout.
func (uid *UserIdentity) ResetTimeout() error {
	return nil
}
