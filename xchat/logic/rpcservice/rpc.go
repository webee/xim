package rpcservice

import (
	"time"
	"xim/xchat/logic/db"
	"xim/xchat/logic/logger"
	"xim/xchat/logic/rpcservice/types"

	ol "github.com/go-ozzo/ozzo-log"
)

// variables
var (
	l *ol.Logger
)

func init() {
	l = logger.Logger.GetLogger("service")
}

// RPCXChat provide xchat rpc services.
type RPCXChat struct {
}

// Echo send msg back.
func (r *RPCXChat) Echo(s string, reply *string) (err error) {
	//l.Info("echo: %s", s)
	time.Sleep(1 * time.Millisecond)
	*reply = s
	return nil
}

// FetchChatMembers fetch chat's members.
func (r *RPCXChat) FetchChatMembers(chatID uint64, reply *[]db.Member) (err error) {
	members, err := db.GetChatMembers(chatID)
	if err != nil {
		return err
	}
	*reply = members
	return nil
}

// SendMsg sends message.
func (r *RPCXChat) SendMsg(args *types.SendMsgArgs, reply **db.Message) (err error) {
	message, err := db.NewMsg(args.ChatID, args.User, args.Msg)
	if err != nil {
		return err
	}

	*reply = message
	return nil
}
