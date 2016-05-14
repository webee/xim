package userboard

import (
	"xim/broker/userds"
	"xim/utils/jwtutils"
)

// VerifyAppAuthToken verify app token.
func VerifyAppAuthToken(auth string) (aid *userds.AppIdentity, err error) {
	token, err := jwtutils.ParseToken(auth, appKey)

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
	token, err := jwtutils.ParseToken(auth, userKey)
	if err != nil || !token.Valid {
		return nil, err
	}
	uid = &userds.UserIdentity{
		App:  token.Claims["app"].(string),
		User: token.Claims["user"].(string),
	}
	return uid, nil
}
