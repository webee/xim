package mid

// UserMsg is a user send message.
type UserMsg struct {
	User string      `json:"user"`
	ID   uint64      `json:"id"`
	Ts   int64       `json:"ts"`
	Msg  interface{} `json:"msg"`
}

// ChatMsgs is chat's messages.
type ChatMsgs struct {
	ChatID uint64    `json:"chat_id"`
	Type   string    `json:"type"`
	Title  string    `json:"title"`
	Msgs   []UserMsg `json:"msgs"`
}

// Chat is a chat.
type Chat struct {
	Type  string `json:"type"`
	ID    uint64 `json:"id"`
	Title string `json:"title"`
	Tag   string `json:"tag"`
}
