package userboard

// UserIdentity is a user instance.
type UserIdentity struct {
	Org  string
	User string
}

// NewUserIdentify creates a new user identity.
func NewUserIdentify(org, user string) *UserIdentity {
	return &UserIdentity{
		Org:  org,
		User: user,
	}
}

// VerifyAuthToken verify user token.
func VerifyAuthToken(token string) (uid *UserIdentity, err error) {
	// TODO http request auth service.
	uid = new(UserIdentity)
	uid.Org = "test"
	uid.User = token
	return uid, nil
}

// IsValid checks if user identity is valid.
func (uid *UserIdentity) IsValid() bool {
	return uid != nil && uid.Org != "" && uid.User != ""
}

// ResetTimeout reset user identity timeout.
func (uid *UserIdentity) ResetTimeout() error {
	return nil
}
