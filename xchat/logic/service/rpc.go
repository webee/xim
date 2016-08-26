package service

import (
	"fmt"
	"xim/xchat/logic/db"
	pubtypes "xim/xchat/logic/pub/types"
	"xim/xchat/logic/service/types"
)

// RPCXChat provide xchat rpc services.
type RPCXChat struct {
}

// Ping is a test rpc method.
func (r *RPCXChat) Ping(args *types.PingArgs, reply *string) (err error) {
	*reply = Ping(args.Sleep, args.Payload)
	return nil
}

// PubUserStatus publish user's status.
func (r *RPCXChat) PubUserStatus(args *types.PubUserStatusArgs, reply *types.NoReply) error {
	return PubUserStatus(args.InstanceID, args.SessionID, args.User, args.Status, args.Info)
}

// SyncOnlineUsers update online users.
func (r *RPCXChat) SyncOnlineUsers(args *types.SyncOnlineUsersArgs, reply *types.NoReply) error {
	for sessionID, user := range args.Users {
		UpdateUserStatus(args.InstanceID, sessionID, user, types.UserStatusOnline)
	}
	return nil
}

// FetchChat fetch chat.
func (r *RPCXChat) FetchChat(chatID uint64, reply *db.Chat) (err error) {
	chat, err := FetchChat(chatID)
	if err != nil {
		return err
	}
	*reply = *chat
	return nil
}

// FetchUserChat fetch user's chat.
func (r *RPCXChat) FetchUserChat(args *types.FetchUserChatArgs, reply *db.UserChat) (err error) {
	userChat, err := FetchUserChat(args.User, args.ChatID)
	if err != nil {
		return err
	}
	*reply = *userChat
	return nil
}

// FetchUserChatList fetch user's chat list.
func (r *RPCXChat) FetchUserChatList(args *types.FetchUserChatListArgs, reply *[]db.UserChat) (err error) {
	userChats, err := FetchUserChatList(args.User, args.OnlyUnsync)
	if err != nil {
		return err
	}
	*reply = userChats
	return nil
}

// SetUserChat set user's chat attribute.
func (r *RPCXChat) SetUserChat(args *types.SetUserChatArgs, reply *types.NoReply) (err error) {
	return SetUserChat(args.User, args.ChatID, args.Key, args.Value)
}

// SyncUserChatRecv sync user's chat msg recv.
func (r *RPCXChat) SyncUserChatRecv(args *types.SyncUserChatRecvArgs, reply *types.NoReply) (err error) {
	return SyncUserChatRecv(args.User, args.ChatID, args.MsgID)
}

// FetchChatMembers fetch chat's members.
func (r *RPCXChat) FetchChatMembers(chatID uint64, reply *[]db.Member) (err error) {
	members, err := FetchChatMembers(chatID)
	if err != nil {
		return err
	}

	*reply = members
	return nil
}

// FetchUserChatMembers fetch user's chat members.
func (r *RPCXChat) FetchUserChatMembers(args *types.FetchUserChatMembersArgs, reply *[]db.Member) (err error) {
	ok, err := IsChatMember(args.ChatID, args.User)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("no permission")
	}

	members, err := FetchChatMembers(args.ChatID)
	if err != nil {
		return err
	}

	*reply = members
	return nil
}

// FetchUserChatMessages fetch chat's messages between sID and eID.
func (r *RPCXChat) FetchUserChatMessages(args *types.FetchUserChatMessagesArgs, reply *[]pubtypes.ChatMessage) (err error) {
	var ok bool

	if args.ChatType != types.ChatTypeRoom {
		// 房间会话不用检查
		ok, err = IsChatMember(args.ChatID, args.User)
		if err != nil {
			return err
		}

		if !ok {
			return fmt.Errorf("no permission")
		}
	}

	msgs, err := FetchChatMessages(args.ChatID, args.ChatType, args.LID, args.RID, args.Limit, args.Desc)
	if err != nil {
		return err
	}
	*reply = msgs
	return nil
}

// FetchChatMessages fetch chat's messages between sID and eID.
func (r *RPCXChat) FetchChatMessages(args *types.FetchChatMessagesArgs, reply *[]pubtypes.ChatMessage) (err error) {
	msgs, err := FetchChatMessages(args.ChatID, args.ChatType, args.LID, args.RID, args.Limit, args.Desc)
	if err != nil {
		return err
	}
	*reply = msgs
	return nil
}

// FetchChatMessagesByIDs fetch chat's messages between by ids.
func (r *RPCXChat) FetchChatMessagesByIDs(args *types.FetchChatMessagesByIDsArgs, reply *[]pubtypes.ChatMessage) (err error) {
	msgs, err := FetchChatMessagesByIDs(args.ChatID, args.ChatType, args.MsgIDs)
	if err != nil {
		return err
	}
	*reply = msgs
	return nil
}

// SendMsg sends message.
func (r *RPCXChat) SendMsg(args *types.SendMsgArgs, reply *pubtypes.ChatMessage) (err error) {
	msg, err := SendChatMsg(args.Source, args.ChatID, args.ChatType, args.Domain, args.User, args.Msg)
	if err != nil {
		return err
	}
	*reply = *msg
	return nil
}

// SendNotify sends notify message.
func (r *RPCXChat) SendNotify(args *types.SendMsgArgs, reply *int64) (err error) {
	ts, err := SendChatNotifyMsg(args.Source, args.ChatID, args.ChatType, args.Domain, args.User, args.Msg)
	if err != nil {
		return err
	}
	*reply = ts
	return nil
}

// SendUserNotify send notify to user.
func (r *RPCXChat) SendUserNotify(args *types.SendUserMsgArgs, reply *int64) error {
	ts, err := SendUserNotify(args.Source, args.ToUser, args.Domain, args.User, args.Msg)
	if err != nil {
		return err
	}
	*reply = ts
	return nil
}

// FetchNewRoomChatIDs fetch room's new chat ids.
func (r *RPCXChat) FetchNewRoomChatIDs(args *types.FetchNewRoomChatIDs, reply *[]uint64) error {
	ids, err := FetchNewRoomChatIDs(args.RoomID, args.ChatIDs)
	if err != nil {
		return err
	}

	*reply = ids
	return nil
}

// JoinChat add user to chat.
func (r *RPCXChat) JoinChat(args *types.JoinExitChatArgs, reply *types.NoReply) error {
	return JoinChat(args.ChatID, args.ChatType, args.User, args.Users)
}

// ExitChat remove user from chat.
func (r *RPCXChat) ExitChat(args *types.JoinExitChatArgs, reply *types.NoReply) error {
	return ExitChat(args.ChatID, args.ChatType, args.User, args.Users)
}
