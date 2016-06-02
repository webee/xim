package main

import (
	"xim/xchat/broker/mid"
	"xim/xchat/broker/router"

	"github.com/valyala/gorpc"
)

func setupMid(xchatRouter *router.XChatRouter, xchatDC *gorpc.DispatcherClient) {
	config := mid.NewConfig(&mid.Config{
		Debug:   args.debug,
		Testing: args.testing,
		Key:     userKey,
	})

	mid.Init()
	mid.Setup(config, xchatRouter, xchatDC)
}
