package router

import (
	"fmt"
	"net/http"
	"strings"
	"time"
	"xim/utils/jwtutils"

	"xim/xchat/broker/logger"

	ol "github.com/go-ozzo/ozzo-log"

	"gopkg.in/webee/turnpike.v2"
)

var (
	l *ol.Logger
)

// Init setup router.
func Init() {
	l = logger.Logger.GetLogger("router")
}

func decodeNSJwt(t string) (ns string, token string) {
	parts := strings.SplitN(t, ":", 2)
	if len(parts) > 1 {
		return parts[0], parts[1]
	}
	return "", t
}

// jwt authentication.
type jwtAuth struct {
	keys map[string][]byte
}

func (e *jwtAuth) Challenge(details map[string]interface{}) (map[string]interface{}, error) {
	l.Debug("challenge: %+v", details)
	return details, nil
}

func (e *jwtAuth) Authenticate(c map[string]interface{}, signature string) (map[string]interface{}, error) {
	l.Debug("Authenticate: %+v", c)

	ns, t := decodeNSJwt(signature)
	key, ok := e.keys[ns]
	if !ok {
		return nil, fmt.Errorf("unknown user namespace: %s", ns)
	}

	claims, err := jwtutils.ParseToken(t, key)
	if err != nil {
		return nil, fmt.Errorf("parse token error: %s", err)
	}
	return map[string]interface{}{"ns": ns, "user": claims["user"].(string), "role": "user"}, nil
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
func NewXChatRouter(userKeys map[string][]byte, debug, testing bool, writeTimeout, pingTimeout, idleTimeout time.Duration) (*XChatRouter, error) {
	if debug {
		turnpike.Debug()
	}

	auth := &jwtAuth{userKeys}
	realms := map[string]turnpike.Realm{
		"xchat": {
			Authorizer:  new(XChatAuthorizer),
			Interceptor: NewDetailsInterceptor(roleIsUser, nil, "details"),
			CRAuthenticators: map[string]turnpike.CRAuthenticator{
				"jwt":    auth,
				"ticket": auth,
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
	s.MaxMsgSize = 64 * 1024
	s.WriteTimeout = writeTimeout
	s.PingTimeout = pingTimeout
	s.IdleTimeout = idleTimeout

	// allow all origins.
	allowAllOrigin := func(r *http.Request) bool { return true }
	s.Upgrader.CheckOrigin = allowAllOrigin

	return &XChatRouter{
		WebsocketServer: s,
	}, nil
}
