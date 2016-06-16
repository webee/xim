package mid

// uris.
const (
	URITestToUpper        = "xchat.test.to_upper"
	URITestAdd            = "xchat.test.add"
	URIWAMPSessionOnJoin  = "wamp.session.on_join"
	URIWAMPSessionOnLeave = "wamp.session.on_leave"

	URIXChatPing = "xchat.ping"

	// 用户信息发布
	URIXChatPubUserInfo = "xchat.user.info.pub"

	// 用户发送消息到会话
	URIXChatSendMsg = "xchat.user.msg.send"

	// 用户发布消息到会话
	URIXChatPubMsg = "xchat.user.msg.pub"

	URIXChatNewChat       = "xchat.user.chat.new"
	URIXChatFetchChat     = "xchat.user.chat.fetch"
	URIXChatFetchChatList = "xchat.user.chat.list"
	URIXChatFetchChatMsgs = "xchat.user.chat.msgs"
	URIXChatSyncChatRecv  = "xchat.user.chat.recv.sync"

	// 用户接收消息
	URIXChatUserMsg = "xchat.user.%d.msg"

	// 房间
	URIXChatEnterRoom = "xchat.user.room.enter"
	URIXChatExitRoom  = "xchat.user.room.exit"

	// 客服
	URIXChatGetCsChat = "xchat.user.cs.chat.get"
)
