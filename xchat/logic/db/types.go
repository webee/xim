package db

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
	return EncodeChatIdentity(ci.Type, ci.ID)
}

// EncodeChatIdentity encode chat_type and chat_id to string chat identity.
func EncodeChatIdentity(chatType string, chatID uint64) string {
	return fmt.Sprintf("%s.%d", chatType, chatID)
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
