package service

import (
	"xim/xchat/logic/db"
	pubtypes "xim/xchat/logic/pub/types"
	"xim/xchat/logic/service/types"
)

// RPCXChat provide xchat rpc services.
type RPCXChat struct {
}

// Echo send msg back.
func (r *RPCXChat) Echo(s string, reply *string) (err error) {
	//l.Info("echo: %s", s)
	//time.Sleep(1 * time.Millisecond)
	*reply = Echo(s)
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

// FetchUserChatList fetch user's chat list.
func (r *RPCXChat) FetchUserChatList(user string, reply *[]db.UserChat) (err error) {
	userChats, err := FetchUserChatList(user)
	if err != nil {
		return err
	}
	*reply = userChats
	return nil
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

// FetchChatMessages fetch chat's messages between sID and eID.
func (r *RPCXChat) FetchChatMessages(args *types.FetchChatMessagesArgs, reply *[]pubtypes.Message) (err error) {
	msgs, err := FetchChatMessages(args.ChatID, args.SID, args.EID)
	if err != nil {
		return err
	}
	*reply = msgs
	return nil
}

// SendMsg sends message.
func (r *RPCXChat) SendMsg(args *types.SendMsgArgs, reply *pubtypes.Message) (err error) {
	msg, err := SendMsg(args.ChatID, args.User, args.Msg)
	if err != nil {
		return err
	}

	*reply = *msg
	return nil
}
