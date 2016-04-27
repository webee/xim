package userboard

import (
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
func VerifyAuthToken(token string) (uid *UserIdentity, err error) {
	// TODO http request auth service.
	uid = &UserIdentity{
		App:  "test",
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
