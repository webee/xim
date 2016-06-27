package service

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// errors.
var (
	ErrBadChatIdentity = errors.New("bad chat identity")
)

// ChatIdentity is a chat with type.
type ChatIdentity struct {
	ID   uint64
	Type string
}

func (ci ChatIdentity) String() string {
	return fmt.Sprintf("%s.%d", ci.Type, ci.ID)
}

// ParseChatIdentity parse chat identity from string.
func ParseChatIdentity(s string) (chatIdentity *ChatIdentity, err error) {
	parts := strings.SplitN(s, ".", 2)
	if len(parts) != 2 {
		return nil, ErrBadChatIdentity
	}
	id, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return nil, err
	}
	return &ChatIdentity{
		ID:   id,
		Type: parts[0],
	}, nil
}
