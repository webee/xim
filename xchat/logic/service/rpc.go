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

// FetchChatMembers fetch chat's members.
func (r *RPCXChat) FetchChatMembers(chatID uint64, reply *[]db.Member) (err error) {
	members, err := FetchChatMembers(chatID)
	if err != nil {
		return err
	}
	*reply = members
	return nil
}

// SendMsg sends message.
func (r *RPCXChat) SendMsg(args *types.SendMsgArgs, reply **pubtypes.Message) (err error) {
	msg, err := SendMsg(args.ChatID, args.User, args.Msg)
	if err != nil {
		return err
	}

	*reply = msg
	return nil
}
