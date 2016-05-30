package db

// Chat is a conversation.
type Chat struct {
	ID      uint64 `db:"id"`
	Type    string
	Channel string `json:"channel,omitempty"`
	Title   string `json:"title,omitempty"`
}

// MemberInfo is chat's member info.
type MemberInfo struct {
	Channel string
	InitID  uint64 `db:"init_id"`
}
