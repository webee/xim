package userboard

// UserMsgBox represents a user msg box.
type UserMsgBox interface {
	PushMsg(v interface{}) (err error)
}
