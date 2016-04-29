package userds

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

// UserLocation represents a user connection location.
type UserLocation struct {
	UserIdentity
	Broker   string
	Instance string
}

func (u UserLocation) String() string {
	return fmt.Sprintf("%s>%s#%s", u.UserIdentity, u.Broker, u.Instance)
}

// IsValid checks if user identity is valid.
func (uid *UserIdentity) IsValid() bool {
	return uid != nil && uid.App != "" && uid.User != ""
}

// ResetTimeout reset user identity timeout.
func (uid *UserIdentity) ResetTimeout() error {
	return nil
}

// ParseUserIdentify parse a user identity from a string.
func ParseUserIdentify(s string) *UserIdentity {
	parts := strings.Split(s, ":")
	return &UserIdentity{
		App:  parts[0],
		User: parts[1],
	}
}

// ParseUserLocation parse a user location from a string.
func ParseUserLocation(s string) *UserLocation {
	parts := strings.Split(s, ">")
	parts2 := strings.Split(parts[1], "#")
	return &UserLocation{
		UserIdentity: *ParseUserIdentify(parts[0]),
		Broker:       parts2[0],
		Instance:     parts2[1],
	}
}
