package userboard

// UserConn represents a user connection(ws/tcp).
type UserConn interface {
	PushMsg(v interface{}) (err error)
}
