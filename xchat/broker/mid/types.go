package mid

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	pubtypes "xim/xchat/logic/pub/types"
)

var (
	ErrBadChatIdentity = errors.New("bad chat identity")
)

// ChatIdentity is a chat with type.
type ChatIdentity struct {
	ID   uint64
	Type string
}

func (ci ChatIdentity) String() string {
	return fmt.Sprintf("%s#%d", ci.Type, ci.ID)
}

// ParseChatIdentity parse chat identity from string.
func ParseChatIdentity(s string) (chatIdentity *ChatIdentity, err error) {
	parts := strings.SplitN(s, "#", 2)
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

// Message is a chat message.
type Message struct {
	ChatID string `json:"chat_id"`
	User   string `json:"user"`
	ID     uint64 `json:"id"`
	Ts     int64  `json:"ts"`
	Msg    string `json:"msg"`
}

// NewMessageFromDBMsg converts db.Message to Message.
func NewMessageFromDBMsg(msg *pubtypes.ChatMessage) *Message {
	return &Message{
		ChatID: fmt.Sprintf("%s#%d", msg.ChatType, msg.ChatID),
		User:   msg.User,
		ID:     msg.ID,
		Ts:     msg.Ts,
		Msg:    msg.Msg,
	}
}

// NotifyMessage is a chat notify message.
type NotifyMessage struct {
	ChatID string `json:"chat_id"`
	User   string `json:"user"`
	Ts     int64  `json:"ts"`
	Msg    string `json:"msg"`
}

// NewNotifyMessageFromDBMsg converts db.Message to Message.
func NewNotifyMessageFromDBMsg(msg *pubtypes.ChatNotifyMessage) *NotifyMessage {
	return &NotifyMessage{
		ChatID: fmt.Sprintf("%s#%d", msg.ChatType, msg.ChatID),
		User:   msg.User,
		Ts:     msg.Ts,
		Msg:    msg.Msg,
	}
}
