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

func sessionDetails(session *turnpike.Session) map[string]interface{} {
	return session.Details
}

// NewSessionDetailsInterceptor returns the default interceptor, which does nothing.
func NewSessionDetailsInterceptor(sessionChecker SessionChecker, sessionTransformer SessionTransformer, key string) turnpike.Interceptor {
	if sessionChecker == nil {
		sessionChecker = sessionOk
	}

	if sessionTransformer == nil {
		sessionTransformer = sessionDetails
	}

	return &DetailsInterceptor{
		sessionChecker:     sessionChecker,
		sessionTransformer: sessionTransformer,
		key:                key,
	}
}

// Intercept do inject work.
func (di *DetailsInterceptor) Intercept(session *turnpike.Session, msg *turnpike.Message) {
	switch (*msg).MessageType() {
	case turnpike.CALL:
		if di.sessionChecker(session) {
			call := (*msg).(*turnpike.Call)
			if call.ArgumentsKw == nil {
				call.ArgumentsKw = make(map[string]interface{})
			}
			call.ArgumentsKw[di.key] = di.sessionTransformer(session)
		}
	case turnpike.PUBLISH:
		if di.sessionChecker(session) {
			publish := (*msg).(*turnpike.Publish)
			if publish.ArgumentsKw == nil {
				publish.ArgumentsKw = make(map[string]interface{})
			}
			publish.ArgumentsKw[di.key] = di.sessionTransformer(session)
		}
	}
}
