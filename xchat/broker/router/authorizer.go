package router

import (
	"fmt"
	"strings"

	"gopkg.in/webee/turnpike.v2"
)

// uris.
const (
	URIXChatUserMsg = "xchat.user.%d.msg"
)

// XChatAuthorizer is authorizor based on user roles.
type XChatAuthorizer struct {
}

// Authorize authorize the reqeust.
func (a *XChatAuthorizer) Authorize(session *turnpike.Session, msg turnpike.Message) (bool, error) {
	role := session.Details["role"]
	if role == "user" {
		switch msg.MessageType() {
		case turnpike.REGISTER:
			return false, nil
		case turnpike.CALL:
			return true, nil
		case turnpike.PUBLISH:
			pub := msg.(*turnpike.Publish)
			topic := string(pub.Topic)
			if strings.HasPrefix(topic, "xchat.user.") && strings.HasSuffix(topic, ".pub") {
				return true, nil
			}
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
