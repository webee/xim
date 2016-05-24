package mid

// uris.
const (
	URITestToUpper        = "xchat.test.to_upper"
	URITestAdd            = "xchat.test.add"
	URIWAMPSessionOnJoin  = "wamp.session.on_join"
	URIWAMPSessionOnLeave = "wamp.session.on_leave"
	URIXChatLogin         = "xchat.user.login"
	URIXChatSendMsg       = "xchat.user.msg.send"
	URIXChatFetchChatList = "xchat.user.chat.list"
	URIXChatFetchChatMsg  = "xchat.user.chat.msg"
	// 用户接收消息
	URIXChatUserMsg = "xchat.user.%d.msg"
	// 用户发送消息的返回
	URIXChatUserReply = "xchat.user.%d.reply"
)
