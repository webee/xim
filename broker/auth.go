package broker

var ()

// UserIdentity is a user instance.
type UserIdentity struct {
	org  string
	user string
}

// VerifyAuthToken verify user token.
func VerifyAuthToken(token string) (uid *UserIdentity, err error) {
	uid = new(UserIdentity)
	uid.org = "test"
	uid.user = "webee"
	return uid, nil
}

// IsValid checks if user identity is valid.
func (uid *UserIdentity) IsValid() bool {
	return uid != nil && uid.org != "" && uid.user != ""
}

// ResetTimeout reset user identity timeout.
func (uid *UserIdentity) ResetTimeout() error {
	return nil
}
