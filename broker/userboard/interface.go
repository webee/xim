package userboard

import "xim/utils/msgutils"

// UserMsgBox represents a user msg box.
type UserMsgBox interface {
	PushMsg(msg msgutils.Message) (err error)
}
