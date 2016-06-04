package types

// SendMsgArgs is the arguments of SendMsg.
type SendMsgArgs struct {
	ChatID uint64
	User   string
	Msg    string
}

// XChatService methods.
const (
	RPCXChatEcho    = "RPCXChat.Echo"
	RPCXChatSendMsg = "RPCXChat.SendMsg"
)
