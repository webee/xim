package router

import (
	"fmt"
	"net/http"
	"xim/utils/jwtutils"

	"xim/xchat/broker/logger"

	ol "github.com/go-ozzo/ozzo-log"

	"gopkg.in/jcelliott/turnpike.v2"
)

var (
	l *ol.Logger
)

// Init setup router.
func Init() {
	l = logger.Logger.GetLogger("router")
}

// jwt authentication.
type jwtAuth struct {
	key   []byte
	csKey []byte
}

func (e *jwtAuth) Challenge(details map[string]interface{}) (map[string]interface{}, error) {
	l.Debug("challenge: %+v", details)
	return details, nil
}

func (e *jwtAuth) Authenticate(c map[string]interface{}, signature string) (map[string]interface{}, error) {
	l.Debug("Authenticate: %+v", c)
	authmethods, ok := c["authmethods"]
	if !ok {
		return nil, fmt.Errorf("no authmethods")
	}
	methods, ok := authmethods.([]interface{})
	if !ok || len(methods) < 1 {
		return nil, fmt.Errorf("bad authmethods")
	}
	method := methods[0]
	switch method {
	case "jwt":
		token, err := jwtutils.ParseToken(signature, e.key)
		if err != nil {
			return nil, fmt.Errorf("parse token error: %s", err)
		}
		return map[string]interface{}{"user": token.Claims["user"], "role": "user"}, nil
	case "cs:jwt":
		token, err := jwtutils.ParseToken(signature, e.key)
		if err != nil {
			return nil, fmt.Errorf("parse token error: %s", err)
		}
		return map[string]interface{}{"user": token.Claims["user"], "role": "user", "app": "cs"}, nil
	}
	return nil, fmt.Errorf("unkown authmethod: %s", method)
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
func NewXChatRouter(userKey, csUserKey []byte, debug, testing bool) (*XChatRouter, error) {
	if debug {
		turnpike.Debug()
	}

	realms := map[string]turnpike.Realm{
		"xchat": {
			Authorizer:  new(XChatAuthorizer),
			Interceptor: NewDetailsInterceptor(roleIsUser, nil, "details"),
			CRAuthenticators: map[string]turnpike.CRAuthenticator{
				"jwt": &jwtAuth{key: userKey, csKey: csUserKey},
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
