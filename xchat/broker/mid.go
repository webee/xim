package main

import (
	"xim/xchat/broker/mid"
	"xim/xchat/broker/router"
)

func setupMid(xchatRouter *router.XChatRouter) {
	config := mid.NewConfig(&mid.Config{
		Debug:        args.debug,
		Testing:      args.testing,
		Key:          userKey,
		LogicRPCAddr: args.logicRPCAddr,
		LogicPubAddr: args.logicPubAddr,
		XChatHostURL: args.XChatHostURL,
	})

	mid.Setup(config, xchatRouter)
}
