package router

import (
	"gopkg.in/webee/turnpike.v2"
)

// SessionChecker check the Session to determine whether do inject.
type SessionChecker func(*turnpike.Session) bool

// SessionTransformer generate details to be injected from session.
type SessionTransformer func(*turnpike.Session) map[string]interface{}

// DetailsInterceptor inject session details for pub and call.
type DetailsInterceptor struct {
	sessionChecker     SessionChecker
	sessionTransformer SessionTransformer
	key                string
}

func sessionOk(session *turnpike.Session) bool {
	return true
}

// SessionIDInterceptor inject session id for pub and call.
type SessionIDInterceptor struct {
	sessionChecker SessionChecker
	key            string
}

func getSessionIDFromSession(session *turnpike.Session) turnpike.ID {
	return session.Id
}

// NewSessionIDInterceptor returns the default interceptor, which does nothing.
func NewSessionIDInterceptor(sessionChecker SessionChecker, key string) turnpike.Interceptor {
	if sessionChecker == nil {
		sessionChecker = sessionOk
	}

	return &SessionIDInterceptor{
		sessionChecker: sessionChecker,
		key:            key,
	}
}

// Intercept do inject work.
func (di *SessionIDInterceptor) Intercept(session *turnpike.Session, msg *turnpike.Message) {
	switch (*msg).MessageType() {
	case turnpike.CALL:
		if di.sessionChecker(session) {
			call := (*msg).(*turnpike.Call)
			if call.ArgumentsKw == nil {
				call.ArgumentsKw = make(map[string]interface{})
			}
			call.ArgumentsKw[di.key] = getSessionIDFromSession(session)
		}
	case turnpike.PUBLISH:
		if di.sessionChecker(session) {
			publish := (*msg).(*turnpike.Publish)
			if publish.ArgumentsKw == nil {
				publish.ArgumentsKw = make(map[string]interface{})
			}
			publish.ArgumentsKw[di.key] = getSessionIDFromSession(session)
		}
	}
}
