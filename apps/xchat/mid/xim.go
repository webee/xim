package mid

import "time"

// XIMClient handles xim apis and app ws connections.
type XIMClient struct {
	// sessionID to uid.
	users map[uint64]uint32
}

// NewXIMClient create a xim client.
func NewXIMClient(config *Config) *XIMClient {
	return nil
}

// Register register user with sessionID.
func (c *XIMClient) Register(sessionID uint64, user string) error {
	time.Sleep(500 * time.Millisecond)
	return nil
}

// Unregister unregister sessionID user.
func (c *XIMClient) Unregister(sessionID uint64) error {
	return nil
}

// Ping pint user.
func (c *XIMClient) Ping(sessionID uint64) error {
	return nil
}

// SendMsg send uid's msg to channel.
func (c *XIMClient) SendMsg(id uint64, sessionID uint64, channel string, msg interface{}) error {
	time.Sleep(200 * time.Millisecond)
	return nil
}
