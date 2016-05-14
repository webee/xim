package main

import (
	"log"
	"net/http"
	"xim/apps/xchat/router"
)

func startRouter(r *router.XChatRouter) {
	go func() {
		httpServeMux := http.NewServeMux()
		if args.debug {
			httpServeMux.Handle("/", http.FileServer(http.Dir(args.testWebDir)))
		}
		httpServeMux.Handle(args.endpoint, r)
		httpServer := &http.Server{
			Handler: httpServeMux,
			Addr:    args.addr,
		}
		log.Println("http listen on: ", args.addr)
		log.Fatalln(httpServer.ListenAndServe())
	}()
}
