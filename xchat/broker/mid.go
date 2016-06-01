package main

import (
	"xim/xchat/broker/mid"
	"xim/xchat/broker/router"
)

func setupMid(xchatRouter *router.XChatRouter) {
	config := mid.NewConfig(
		&mid.Config{
			Debug:   args.debug,
			Testing: args.testing,
			Key:     userKey,
		})

	mid.Setup(config, xchatRouter)
}
