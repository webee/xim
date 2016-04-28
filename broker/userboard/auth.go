package userboard

import (
	"errors"
	"fmt"
	"strings"
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
func VerifyAuthToken(app, token string) (uid UserIdentity, err error) {
	if app != "test" {
		err = errors.New("bad token")
		return
	}
	// TODO http request auth service.
	uid = UserIdentity{
		App:  app,
		User: token,
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
