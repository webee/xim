package mid

import (
	"fmt"
	"time"
	pubtypes "xim/xchat/logic/pub/types"
)

// Timestamp is timestamp.
type Timestamp time.Time

// MarshalJSON encode Timestamp to byte array.
func (t Timestamp) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%d", time.Time(t).Unix())), nil
}

// Message is chat message.
type Message struct {
	ChatID uint64 `json:"chat_id"`
	User   string `json:"user"`
	MsgID  uint64 `json:"msg_id"`
	Ts     int64  `json:"ts"`
	Msg    string `json:"msg"`
}

// NewMessageFromDBMsg converts db.Message to Message.
func NewMessageFromDBMsg(message *pubtypes.Message) *Message {
	return &Message{
		ChatID: message.ChatID,
		User:   message.User,
		MsgID:  message.MsgID,
		Ts:     message.Ts,
		Msg:    message.Msg,
	}
}
