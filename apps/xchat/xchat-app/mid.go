package main

import (
	"xim/apps/xchat/mid"
	"xim/apps/xchat/router"
)

func setupMid(xchatRouter *router.XChatRouter) {
	config := mid.NewConfig(
		&mid.Config{
			Debug:       args.debug,
			XIMHostURL:  args.ximHostURL,
			XIMApp:      args.ximApp,
			XIMPassword: args.ximPassword,
			XIMAppWsURL: args.ximAppWsURL,
		})

	mid.Setup(config, xchatRouter)
}
