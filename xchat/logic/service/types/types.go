package types

// chat type
const (
	ChatTypeRoom  = "room"
	ChatTypeSelf  = "self"
	ChatTypeUser  = "user"
	ChatTypeGroup = "group"
	ChatTypeCS    = "cs"
)

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

// PingArgs is the arguments of Ping.
type PingArgs struct {
	Sleep   int64
	Payload string
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

// PubUserStatusArgs is the arguments of PubUserStatus
type PubUserStatusArgs struct {
	InstanceID uint64
	SessionID  uint64
	User       string
	Status     string
	Info       string
}

// SyncOnlineUsersArgs is the arguments of SyncOnlineUsers
type SyncOnlineUsersArgs struct {
	InstanceID uint64
	Users      map[uint64]string
}

// FetchNewRoomChatIDs is the arguments of FetchNewRoomChatIDs
type FetchNewRoomChatIDs struct {
	RoomID  uint64
	ChatIDs []uint64
}

// XChatService methods.
const (
	RPCXChatPing                  = "RPCXChat.Ping"
	RPCXChatPubUserStatus         = "RPCXChat.PubUserStatus"
	RPCXChatSyncOnlineUsers       = "RPCXChat.SyncOnlineUsers"
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
