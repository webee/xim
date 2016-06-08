package types

// msg kinds.
const (
	MsgKindChat       = "chat"
	MsgKindChatNotify = "chat_notify"
)

// SendMsgArgs is the arguments of SendMsg.
type SendMsgArgs struct {
	ChatID uint64
	User   string
	Msg    string
	Kind   string
}

// FetchChatMessagesArgs is the arguments of FetchChatMessages
type FetchChatMessagesArgs struct {
	ChatID uint64
	SID    uint64
	EID    uint64
}

// XChatService methods.
const (
	RPCXChatEcho              = "RPCXChat.Echo"
	RPCXChatSendMsg           = "RPCXChat.SendMsg"
	RPCXChatFetchChatMessages = "RPCXChat.FetchChatMessages"
	RPCXChatFetchChat         = "RPCXChat.FetchChat"
	RPCXChatFetchUserChatList = "RPCXChat.FetchUserChatList"
	RPCXChatFetchChatMembers  = "RPCXChat.FetchChatMembers"
)
