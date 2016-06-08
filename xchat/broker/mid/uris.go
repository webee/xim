package mid

// uris.
const (
	URITestToUpper        = "xchat.test.to_upper"
	URITestAdd            = "xchat.test.add"
	URIWAMPSessionOnJoin  = "wamp.session.on_join"
	URIWAMPSessionOnLeave = "wamp.session.on_leave"

	URIXChatPing = "xchat.ping"

	// biz
	URIXChatSendMsg       = "xchat.user.msg.send"
	URIXChatFetchChatList = "xchat.user.chat.list"
	URIXChatFetchChatMsgs = "xchat.user.chat.msgs"

	// 用户发布消息
	URIXChatUserPub = "xchat.user.chat.pub"

	URIXChatNewChat   = "xchat.user.chat.new"
	URIXChatFetchChat = "xchat.user.chat.fetch"
	URIXChatChatList  = "xchat.user.chat.list"

	// 用户接收消息
	URIXChatUserMsg = "xchat.user.%d.msg"

	// 房间
	URIXChatEnterRoom = "xchat.user.room.enter"
	URIXChatExitRoom  = "xchat.user.room.exit"

	// 客服
	URIXChatGetCsChat = "xchat.user.cs.chat.get"
)
