package logic

import "fmt"

// UserLocation represents a user connection location.
type UserLocation struct {
	Broker   string
	Org      string
	User     string
	Instance string
}

func (u UserLocation) String() string {
	return fmt.Sprintf("%s:%s#%s>%s", u.Org, u.User, u.Instance, u.Broker)
}
