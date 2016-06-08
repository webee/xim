package types

// msg kinds.
const (
	MsgKindChat       = "chat"
	MsgKindChatNotify = "chat_notify"
)

// NoArgs used by rpc with no args.
type NoArgs struct {
}

// NoReply used by rpc with no reply.
type NoReply struct {
}

// SendMsgArgs is the arguments of SendMsg.
type SendMsgArgs struct {
	ChatID uint64
	User   string
	Msg    string
	Kind   string
}

// FetchUserChatArgs is the arguments of FetchUserChat
type FetchUserChatArgs struct {
	User   string
	ChatID uint64
}

// FetchUserChatListArgs is the arguments of FetchUserChatList
type FetchUserChatListArgs struct {
	User       string
	OnlyUnsync bool
}

// SyncUserChatRecvArgs is the arguments of SyncUserChatRecv.
type SyncUserChatRecvArgs struct {
	User   string
	ChatID uint64
	MsgID  uint64
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
	RPCXChatFetchUserChat     = "RPCXChat.FetchUserChat"
	RPCXChatFetchUserChatList = "RPCXChat.FetchUserChatList"
	RPCXChatSyncUserChatRecv  = "RPCXChat.SyncUserChatRecv"
	RPCXChatFetchChatMembers  = "RPCXChat.FetchChatMembers"
)
