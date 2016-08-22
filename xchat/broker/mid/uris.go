package mid

// uris.
const (
	URIWAMPSessionOnJoin  = "wamp.session.on_join"
	URIWAMPSessionOnLeave = "wamp.session.on_leave"

	URIXChatPing = "xchat.ping"

	// 用户信息发布
	URIXChatPubUserInfo       = "xchat.user.info.pub"
	URIXChatPubUserStatusInfo = "xchat.user.status.pub"

	// 用户发送消息到会话
	URIXChatSendMsg = "xchat.user.msg.send"

	// 用户发送notify消息到会话
	URIXChatSendNotify = "xchat.user.notify.send"
	URIXChatPubNotify  = "xchat.user.notify.pub"

	// 用户发送notify消息到用户
	URIXChatSendUserNotify = "xchat.user.usernotify.send"
	URIXChatPubUserNotify  = "xchat.user.usernotify.pub"

	URIXChatNewChat          = "xchat.user.chat.new"
	URIXChatFetchChat        = "xchat.user.chat.fetch"
	URIXChatFetchChatList    = "xchat.user.chat.list"
	URIXChatFetchChatMembers = "xchat.user.chat.members"
	URIXChatFetchChatMsgs    = "xchat.user.chat.msgs"
	URIXChatSetChat          = "xchat.user.chat.set"
	URIXChatSyncChatRecv     = "xchat.user.chat.recv.sync"

	// 会话
	URIXChatJoinChat = "xchat.user.chat.join"
	URIXChatExitChat = "xchat.user.chat.exit"

	// 用户接收消息
	URIXChatUserMsg = "xchat.user.%d.msg"

	// 房间
	URIXChatEnterRoom = "xchat.user.room.enter"
	URIXChatExitRoom  = "xchat.user.room.exit"

	// 客服
	URIXChatGetCsChat = "xchat.user.cs.chat.get"
)
