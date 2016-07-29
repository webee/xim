package router

import (
	"gopkg.in/webee/turnpike.v2"
)

// DetailsChecker check the details to determine whether do inject.
type DetailsChecker func(map[string]interface{}) bool

// DetailsTransformer generate details to be injected from session details.
type DetailsTransformer func(map[string]interface{}) map[string]interface{}

// DetailsInterceptor inject session details for pub and call.
type DetailsInterceptor struct {
	detailsChecker     DetailsChecker
	detailsTransformer DetailsTransformer
	key                string
}

func detailsOk(details map[string]interface{}) bool {
	return true
}

func identityDetails(details map[string]interface{}) map[string]interface{} {
	return details
}

// NewDetailsInterceptor returns the default interceptor, which does nothing.
func NewDetailsInterceptor(detailsChecker DetailsChecker, detailsTransformer DetailsTransformer, key string) turnpike.Interceptor {
	if detailsChecker == nil {
		detailsChecker = detailsOk
	}

	if detailsTransformer == nil {
		detailsTransformer = identityDetails
	}

	return &DetailsInterceptor{
		detailsChecker:     detailsChecker,
		detailsTransformer: detailsTransformer,
		key:                key,
	}
}

// Intercept do inject work.
func (di *DetailsInterceptor) Intercept(session *turnpike.Session, msg *turnpike.Message) {
	switch (*msg).MessageType() {
	case turnpike.CALL:
		if di.detailsChecker(session.Details) {
			call := (*msg).(*turnpike.Call)
			if call.ArgumentsKw == nil {
				call.ArgumentsKw = make(map[string]interface{})
			}
			call.ArgumentsKw[di.key] = di.detailsTransformer(session.Details)
		}
	case turnpike.PUBLISH:
		if di.detailsChecker(session.Details) {
			publish := (*msg).(*turnpike.Publish)
			if publish.ArgumentsKw == nil {
				publish.ArgumentsKw = make(map[string]interface{})
			}
			publish.ArgumentsKw[di.key] = di.detailsTransformer(session.Details)
		}
	}
}
