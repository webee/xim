package userds

import (
	"fmt"
	"strings"
)

// AppIdentity is a app instance.
type AppIdentity struct {
	App string
}

func (aid *AppIdentity) String() string {
	return fmt.Sprintf("%s", aid.App)
}

// AppLocation represents a app connection location.
type AppLocation struct {
	AppIdentity
	Broker   string
	Instance string
}

func (a *AppLocation) String() string {
	return fmt.Sprintf("%s>%s#%s", &a.AppIdentity, a.Broker, a.Instance)
}

// UserIdentity is a user instance.
type UserIdentity struct {
	App  string
	User string
}

func (uid *UserIdentity) String() string {
	return fmt.Sprintf("%s:%s", uid.App, uid.User)
}

// IsValid checks if user identity is valid.
func (uid *UserIdentity) IsValid() bool {
	return uid != nil && uid.App != "" && uid.User != ""
}

// UserLocation represents a user connection location.
type UserLocation struct {
	UserIdentity
	Broker   string
	Instance string
}

func (u *UserLocation) String() string {
	return fmt.Sprintf("%s>%s#%s", &u.UserIdentity, u.Broker, u.Instance)
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
