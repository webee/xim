package router

import (
	"net/http"
	"time"

	"xim/xchat/broker/logger"

	ol "github.com/go-ozzo/ozzo-log"

	"gopkg.in/webee/turnpike.v2"
)

var (
	l        *ol.Logger
	emptyMap = map[string]interface{}{}
)

// Init setup router.
func Init() {
	l = logger.Logger.GetLogger("router")
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
	realms map[string]*turnpike.Realm
}

// NewXChatRouter creates a xchat router.
func NewXChatRouter(userKeys map[string][]byte, debug, testing bool, writeTimeout, pingTimeout, idleTimeout time.Duration) (*XChatRouter, error) {
	if debug {
		turnpike.Debug()
	}

	auth := &jwtAuth{userKeys}
	xauth := &xjwtAuth{userKeys}
	realms := map[string]*turnpike.Realm{
		"xchat": {
			Authorizer:  new(XChatAuthorizer),
			Interceptor: NewDetailsInterceptor(roleIsUser, nil, "details"),
			CRAuthenticators: map[string]turnpike.CRAuthenticator{
				"jwt":    auth,
				"ticket": auth,
			},
			Authenticators: map[string]turnpike.Authenticator{
				"xjwt": xauth,
			},
		},
	}
	if testing {
		realms["realm1"] = &turnpike.Realm{}
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
		realms:          realms,
	}, nil
}

// GetRealm get specified realm.
func (r *XChatRouter) GetRealm(name string) (*turnpike.Realm, bool) {
	if realm, ok := r.realms[name]; ok {
		return realm, ok
	}
	return nil, false
}
