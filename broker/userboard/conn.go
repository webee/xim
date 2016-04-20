package userboard

// MsgBroker represents a broker(ws/tcp).
type MsgBroker interface {
	WriteMsg(v interface{}) (err error)
}
