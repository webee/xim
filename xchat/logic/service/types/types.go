package types

import (
	pubtypes "xim/xchat/logic/pub/types"
)

// chat type
const (
	ChatTypeRoom  = "room"
	ChatTypeSelf  = "self"
	ChatTypeUser  = "user"
	ChatTypeUsers = "users"
	ChatTypeGroup = "group"
	ChatTypeCS    = "cs"
)

// msg kinds.
const (
	MsgKindChat       = "chat"
	MsgKindChatNotify = "chat_notify"
	MsgKindUserNotify = "user_notify"
	MsgKindUserSysReq = "user_sys_req"
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

// SendMsgOptions is options for send msgs.
type SendMsgOptions struct {
	IgnorePermCheck     bool
	IgnoreNotifyOffline bool
	IgnoreMsgNotify     bool
}

// SendMsgArgs is the arguments of SendMsg.
type SendMsgArgs struct {
	Source           *pubtypes.MsgSource
	ChatID           uint64
	ChatType         string
	Domain           string
	User             string
	Msg              string
	ForceNotifyUsers map[string]struct{}
	Options          *SendMsgOptions
}

// SendUserMsgArgs is the arguments of SendUserMsg.
type SendUserMsgArgs struct {
	Source  *pubtypes.MsgSource
	User    string
	ToUser  string
	Domain  string
	Msg     string
	Options *SendMsgOptions
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
	LastMsgTs  int64
}

// SetUserChatArgs is the arguments of SetUserChat.
type SetUserChatArgs struct {
	User   string
	ChatID uint64
	Key    string
	Value  interface{}
}

// SetChatTitleArgs is the arguments of SetChatTitle.
type SetChatTitleArgs struct {
	User     string
	ChatID   uint64
	ChatType string
	Title    string
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

// FetchChatMessagesByIDsArgs is the arguments of FetchChatMessagesByIDs
type FetchChatMessagesByIDsArgs struct {
	ChatID   uint64
	ChatType string
	MsgIDs   []uint64
}

// FetchUserChatMembersArgs is the arguments of FetchUserChatMembers
type FetchUserChatMembersArgs struct {
	User   string
	ChatID uint64
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
}

// PubUserInfoArgs is the arguments of PubUserInfo
type PubUserInfoArgs struct {
	PubUserStatusArgs
	Info string
}

// SyncOnlineUsersArgs is the arguments of SyncOnlineUsers
type SyncOnlineUsersArgs struct {
	InstanceID uint64
	Users      map[uint64]string
}

// JoinExitChatArgs is the arguments of Join(Exit)Chat.
type JoinExitChatArgs struct {
	ChatID   uint64
	ChatType string
	User     string
	Users    []string
}

// FetchNewRoomChatsArgs is the arguments of FetchNewRoomChatIDs
type FetchNewRoomChatsArgs struct {
	RoomID  uint64
	ChatIDs []uint64
}

// XChatService methods.
const (
	RPCXChatPing                   = "RPCXChat.Ping"
	RPCXChatPubUserStatus          = "RPCXChat.PubUserStatus"
	RPCXChatPubUserInfo            = "RPCXChat.PubUserInfo"
	RPCXChatSyncOnlineUsers        = "RPCXChat.SyncOnlineUsers"
	RPCXChatSendMsg                = "RPCXChat.SendMsg"
	RPCXChatSendNotify             = "RPCXChat.SendNotify"
	RPCXChatFetchChatMessages      = "RPCXChat.FetchChatMessages"
	RPCXChatFetchChatMessagesByIDs = "RPCXChat.FetchChatMessagesByIDs"
	RPCXChatFetchUserChatMessages  = "RPCXChat.FetchUserChatMessages"
	RPCXChatRoomExists             = "RPCXChat.RoomExists"
	RPCXChatFetchChat              = "RPCXChat.FetchChat"
	RPCXChatFetchUserChat          = "RPCXChat.FetchUserChat"
	RPCXChatFetchUserChatList      = "RPCXChat.FetchUserChatList"
	RPCXChatSetUserChat            = "RPCXChat.SetUserChat"
	RPCXChatSyncUserChatRecv       = "RPCXChat.SyncUserChatRecv"
	RPCXChatFetchChatMembers       = "RPCXChat.FetchChatMembers"
	RPCXChatFetchUserChatMembers   = "RPCXChat.FetchUserChatMembers"
	RPCXChatFetchNewRoomChats      = "RPCXChat.FetchNewRoomChats"
	RPCXChatJoinChat               = "RPCXChat.JoinChat"
	RPCXChatExitChat               = "RPCXChat.ExitChat"

	// set chat
	RPCXChatSetChatTitle = "RPCXChat.SetChatTitle"

	// notify
	RPCXChatSendUserNotify = "RPCXChat.SendUserNotify"
)
