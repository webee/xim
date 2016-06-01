package router

import (
	"fmt"
	"log"
	"net/http"
	"xim/utils/jwtutils"

	"gopkg.in/jcelliott/turnpike.v2"
)

// jwt authentication.
type jwtAuth struct {
	key []byte
}

func (e *jwtAuth) Challenge(details map[string]interface{}) (map[string]interface{}, error) {
	log.Println("challenge:", details)
	return details, nil
}

func (e *jwtAuth) Authenticate(c map[string]interface{}, signature string) (map[string]interface{}, error) {
	log.Println("Authenticate:", c)
	token, err := jwtutils.ParseToken(signature, e.key)
	if err != nil {
		return nil, fmt.Errorf("parse token error: %s", err)
	}
	return map[string]interface{}{"user": token.Claims["user"], "role": "user"}, nil
}

func roleIsUser(details map[string]interface{}) bool {
	if val, ok := details["role"]; ok {
		if role, ok := val.(string); ok {
			return role == "user"
		}
	}
	return false
}

var client *turnpike.Client
var realm1 *turnpike.Client

// XChatRouter is a wamp router for xchat.
type XChatRouter struct {
	*turnpike.WebsocketServer
}

// NewXChatRouter creates a xchat router.
func NewXChatRouter(userKey []byte, debug, testing bool) (*XChatRouter, error) {
	if debug {
		turnpike.Debug()
	}
	realms := map[string]turnpike.Realm{
		"xchat": {
			Authorizer:  new(XChatAuthorizer),
			Interceptor: NewDetailsInterceptor(roleIsUser, nil, "details"),
			CRAuthenticators: map[string]turnpike.CRAuthenticator{
				"jwt": &jwtAuth{key: userKey},
			},
		},
	}
	if testing {
		realms["realm1"] = turnpike.Realm{}
	}

	s, err := turnpike.NewWebsocketServer(realms)
	if err != nil {
		return nil, err
	}

	// allow all origins.
	allowAllOrigin := func(r *http.Request) bool { return true }
	s.Upgrader.CheckOrigin = allowAllOrigin

	return &XChatRouter{
		WebsocketServer: s,
	}, nil
}
