package main

import (
	"log"
	"net/http"
	"xim/xchat/broker/router"
)

func startRouter(r *router.XChatRouter) {
	go func() {
		httpServeMux := http.NewServeMux()
		httpServeMux.Handle(args.endpoint, r)
		httpServer := &http.Server{
			Handler: httpServeMux,
			Addr:    args.addr,
		}
		l.Info("http listen on: %s", args.addr)
		log.Fatalln(httpServer.ListenAndServe())
	}()
}
