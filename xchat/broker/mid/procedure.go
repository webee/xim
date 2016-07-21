package mid

import (
	"gopkg.in/webee/turnpike.v2"
)

// Procedure is a simple procudure.
type Procedure func(args []interface{}, kwargs map[string]interface{}) (rargs []interface{}, rkwargs map[string]interface{}, rerr APIError)

// SessionProcedure is user session procudure.
type SessionProcedure func(s *Session, args []interface{}, kwargs map[string]interface{}) (rargs []interface{}, rkwargs map[string]interface{}, rerr APIError)

func (p Procedure) registerTo(client *turnpike.Client, uri string) error {
	return client.BasicRegister(uri, callProcedure(uri, p, false))
}

func (p Procedure) xregisterTo(client *turnpike.Client, uri string) error {
	return client.BasicRegister(uri, callProcedure(uri, p, true))
}

func (p SessionProcedure) registerTo(client *turnpike.Client, uri string) error {
	return client.BasicRegister(uri, callProcedure(uri, p.procedure(), false))
}

func (p SessionProcedure) xregisterTo(client *turnpike.Client, uri string) error {
	return client.BasicRegister(uri, callProcedure(uri, p.procedure(), true))
}

func (p SessionProcedure) procedure() Procedure {
	return func(args []interface{}, kwargs map[string]interface{}) ([]interface{}, map[string]interface{}, APIError) {
		s := getSessionFromDetails(kwargs["details"], false)
		if s == nil {
			return nil, nil, SessionExceptionError
		}
		return p(s, args, kwargs)
	}
}

func callProcedure(uri string, procedure Procedure, logInfo bool) turnpike.BasicMethodHandler {
	return func(args []interface{}, kwargs map[string]interface{}) (result *turnpike.CallResult) {
		defer func() {
			if r := recover(); r != nil {
				l.Warning("[rpc]%s: call error, %s", uri, r)
				result = &turnpike.CallResult{Args: APIErrorToRPCResult(InvalidArgumentError)}
			}
		}()
		if logInfo {
			l.Info("[rpc]%s: %v, %+v\n", uri, args, kwargs)
		} else {
			l.Debug("[rpc]%s: %v, %+v\n", uri, args, kwargs)
		}
		rargs, rkwargs, rerr := procedure(args, kwargs)
		if rerr != nil {
			return &turnpike.CallResult{Args: APIErrorToRPCResult(rerr)}
		}
		return &turnpike.CallResult{Args: rargs, Kwargs: rkwargs}
	}
}
