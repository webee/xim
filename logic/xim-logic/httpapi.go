package main

import (
	"xim/logic/httpapi"
)

func startHTTPApiServer() {
	config := httpapi.NewServerConfig(
		&httpapi.ServerConfig{
			Debug:       args.debug,
			Addr:        args.addr,
			SaltPath:    args.saltPath,
			AppKeyPath:  args.appKeyPath,
			UserKeyPath: args.userKeyPath,
		})
	go httpapi.Start(config)
}
