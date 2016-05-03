package main

import (
	"xim/logic/httpapi"
)

func startHTTPApiServer() {
	config := httpapi.NewServerConfig(
		&httpapi.ServerConfig{
			Debug:       args.debug,
			Addr:        args.addr,
			AppKeyPath:  args.appKeyPath,
			UserKeyPath: args.userKeyPath,
		})
	go httpapi.Start(config)
}
