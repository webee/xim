package userds

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/youtube/vitess/go/pools"
)

var (
	idPool    = pools.NewIDPool()
	appIDPool = pools.NewIDPool()
)

// AppIdentity is a app instance.
type AppIdentity struct {
	App string
}

func (aid AppIdentity) String() string {
	return fmt.Sprintf("%s", aid.App)
}

// AppLocation represents a app connection location.
type AppLocation struct {
	AppIdentity
	Broker   string
	Instance uint32
}

func (a AppLocation) String() string {
	return fmt.Sprintf("%s>%s#%d", &a.AppIdentity, a.Broker, a.Instance)
}

// NewAppLocation create a new app location.
func NewAppLocation(aid *AppIdentity, broker string) *AppLocation {
	return &AppLocation{
		AppIdentity: *aid,
		Broker:      broker,
		Instance:    appIDPool.Get(),
	}
}

// Close return the instance id.
func (a *AppLocation) Close() {
	appIDPool.Put(a.Instance)
}

// UserIdentity is a user instance.
type UserIdentity struct {
	App  string
	User string
}

func (uid UserIdentity) String() string {
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
	Instance uint32
}

// NewUserLocation create a new user location.
func NewUserLocation(uid *UserIdentity, broker string) *UserLocation {
	return &UserLocation{
		UserIdentity: *uid,
		Broker:       broker,
		Instance:     idPool.Get(),
	}
}

// Close return the instance id.
func (u *UserLocation) Close() {
	idPool.Put(u.Instance)
}

func (u UserLocation) String() string {
	return fmt.Sprintf("%s>%s#%d", &u.UserIdentity, u.Broker, u.Instance)
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
	instance, _ := strconv.ParseUint(parts2[1], 10, 32)
	return &UserLocation{
		UserIdentity: *ParseUserIdentify(parts[0]),
		Broker:       parts2[0],
		Instance:     uint32(instance),
	}
}
