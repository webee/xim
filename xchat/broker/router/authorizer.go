package router

import (
	"fmt"

	"github.com/webee/turnpike"
)

// uris.
const (
	URIXChatUserMsg = "xchat.user.%d.msg"
)

// XChatAuthorizer is authorizor based on user roles.
type XChatAuthorizer struct {
}

// Authorize authorize the reqeust.
func (a *XChatAuthorizer) Authorize(session turnpike.Session, msg turnpike.Message) (bool, error) {
	role := session.Details["role"]
	if role == "user" {
		switch msg.MessageType() {
		case turnpike.REGISTER:
			return false, nil
		case turnpike.CALL:
			return true, nil
		case turnpike.PUBLISH:
			return false, nil
		case turnpike.SUBSCRIBE:
			sub := msg.(*turnpike.Subscribe)
			if string(sub.Topic) == fmt.Sprintf(URIXChatUserMsg, session.Id) {
				return true, nil
			}
			return false, nil
		}
	}
	return true, nil
}
