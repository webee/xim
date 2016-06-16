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
	ChatID   uint64
	ChatType string
	User     string
	Msg      string
	Kind     string
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

// FetchUserChatMessagesArgs is the arguments of FetchUserChatMessages
type FetchUserChatMessagesArgs struct {
	User     string
	ChatID   uint64
	ChatType string
	LID      uint64
	RID      uint64
	Limit    int
	Desc     bool
}

// FetchChatMessagesArgs is the arguments of FetchChatMessages
type FetchChatMessagesArgs struct {
	ChatID   uint64
	ChatType string
	LID      uint64
	RID      uint64
	Limit    int
	Desc     bool
}

// user status
const (
	UserStatusOnline  = "online"
	UserStatusOffline = "offline"
)

// PubUserStatusArgs is the arguments of PubUserSTatus
type PubUserStatusArgs struct {
	User   string
	Status string
	Info   string
}

// FetchNewRoomChatIDs is the arguments of FetchNewRoomChatIDs
type FetchNewRoomChatIDs struct {
	RoomID  uint64
	ChatIDs []uint64
}

// XChatService methods.
const (
	RPCXChatEcho                  = "RPCXChat.Echo"
	RPCXChatPubUserStatus         = "RPCXChat.PubUserStatus"
	RPCXChatSendMsg               = "RPCXChat.SendMsg"
	RPCXChatFetchChatMessages     = "RPCXChat.FetchChatMessages"
	RPCXChatFetchUserChatMessages = "RPCXChat.FetchUserChatMessages"
	RPCXChatFetchChat             = "RPCXChat.FetchChat"
	RPCXChatFetchUserChat         = "RPCXChat.FetchUserChat"
	RPCXChatFetchUserChatList     = "RPCXChat.FetchUserChatList"
	RPCXChatSyncUserChatRecv      = "RPCXChat.SyncUserChatRecv"
	RPCXChatFetchChatMembers      = "RPCXChat.FetchChatMembers"
	RPCXChatFetchNewRoomChatIDs   = "RPCXChat.FetchNewRoomChatIDs"
)
