package router

import (
	"gopkg.in/webee/turnpike.v2"
)

// SessionIDInterceptor inject session id for pub and call.
type SessionIDInterceptor struct {
	detailsChecker DetailsChecker
	key            string
}

func getSessionIDFromDetails(details map[string]interface{}) interface{} {
	return details["session"]
}

// NewSessionIDInterceptor returns the default interceptor, which does nothing.
func NewSessionIDInterceptor(detailsChecker DetailsChecker, key string) turnpike.Interceptor {
	if detailsChecker == nil {
		detailsChecker = detailsOk
	}

	return &SessionIDInterceptor{
		detailsChecker: detailsChecker,
		key:            key,
	}
}

// Intercept do inject work.
func (di *SessionIDInterceptor) Intercept(session *turnpike.Session, msg *turnpike.Message) {
	switch (*msg).MessageType() {
	case turnpike.CALL:
		if di.detailsChecker(session.Details) {
			call := (*msg).(*turnpike.Call)
			if call.ArgumentsKw == nil {
				call.ArgumentsKw = make(map[string]interface{})
			}
			call.ArgumentsKw[di.key] = getSessionIDFromDetails(session.Details)
		}
	case turnpike.PUBLISH:
		if di.detailsChecker(session.Details) {
			publish := (*msg).(*turnpike.Publish)
			if publish.ArgumentsKw == nil {
				publish.ArgumentsKw = make(map[string]interface{})
			}
			publish.ArgumentsKw[di.key] = getSessionIDFromDetails(session.Details)
		}
	}
}
