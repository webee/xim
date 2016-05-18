package router

import "github.com/jcelliott/turnpike"

// UserRole defines user roles.
type UserRole struct {
	Publisher  bool
	Subscriber bool
	Callee     bool
	Caller     bool
}

// UserRoleAuthorizer is authorizor based on user roles.
type UserRoleAuthorizer struct {
	roles map[interface{}]UserRole
}

// NewUserRoleAuthorizer returns the user role authorizer struct
func NewUserRoleAuthorizer(roles map[interface{}]UserRole) turnpike.Authorizer {
	return &UserRoleAuthorizer{roles}
}

// Authorize authorize the reqeust.
func (a *UserRoleAuthorizer) Authorize(session turnpike.Session, msg turnpike.Message) (bool, error) {
	role := session.Details["role"]
	if userRole, ok := a.roles[role]; ok {
		switch msg.MessageType() {
		case turnpike.REGISTER:
			return userRole.Callee, nil
		case turnpike.CALL:
			return userRole.Callee, nil
		case turnpike.PUBLISH:
			return userRole.Publisher, nil
		case turnpike.SUBSCRIBE:
			return userRole.Subscriber, nil
		}
	}
	return true, nil
}
