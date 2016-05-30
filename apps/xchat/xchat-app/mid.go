package main

import (
	"xim/apps/xchat/mid"
	"xim/apps/xchat/router"
)

func setupMid(xchatRouter *router.XChatRouter) {
	config := mid.NewConfig(
		&mid.Config{
			Debug:        args.debug,
			Testing:      args.testing,
			XIMHostURL:   args.XimArgs.HostURL,
			XIMApp:       args.XimArgs.App,
			XIMPassword:  args.XimArgs.Password,
			XIMAppWsURL:  args.XimArgs.AppWsURL,
			Key:          userKey,
			XChatHostURL: args.XChatHostURL,
		})

	mid.Setup(config, xchatRouter)
}
