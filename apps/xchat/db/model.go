package db

// Chat is a conversation.
type Chat struct {
	ID      uint64 `db:"id"`
	Type    string
	Channel string
}
