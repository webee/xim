package router

import (
	"fmt"
	"xim/utils/jwtutils"
)

// jwt authentication.
type jwtAuth struct {
	keys map[string][]byte
}

func (e *jwtAuth) Challenge(details map[string]interface{}) (map[string]interface{}, error) {
	l.Debug("challenge: %+v", details)
	return emptyMap, nil
}

func authenticate(keys map[string][]byte, signature string) (map[string]interface{}, error) {
	ns, t := jwtutils.DecodeNSJwt(signature)
	key, ok := keys[ns]
	if !ok {
		return nil, fmt.Errorf("unknown user namespace: %s", ns)
	}

	claims, err := jwtutils.ParseToken(t, key)
	if err != nil {
		return nil, fmt.Errorf("parse token error: %s", err)
	}
	if claims["name"] == nil {
		return nil, fmt.Errorf("bad token")
	}
	return map[string]interface{}{"ns": ns, "name": claims["name"].(string), "role": "user"}, nil
}

func (e *jwtAuth) Authenticate(c map[string]interface{}, signature string) (map[string]interface{}, error) {
	l.Debug("Authenticate: %+v", c)

	return authenticate(e.keys, signature)
}

type xjwtAuth struct {
	keys map[string][]byte
}

func (e *xjwtAuth) Authenticate(details map[string]interface{}) (map[string]interface{}, error) {
	l.Debug("Authenticate: %+v", details)

	authid, ok := details["authid"]
	if !ok {
		return nil, fmt.Errorf("no authid")
	}

	signature, ok := authid.(string)
	if !ok {
		return nil, fmt.Errorf("invalid authid")
	}

	return authenticate(e.keys, signature)
}
