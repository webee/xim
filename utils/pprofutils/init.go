package pprofutils

import (
	"log"
	"net/http"
	"net/http/pprof"
)

// StartPProfListen start debug pprof listening.
func StartPProfListen(addr string) {
	serverMux := http.NewServeMux()
	serverMux.HandleFunc("/debug/pprof/", pprof.Index)
	serverMux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	serverMux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	serverMux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	serverMux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	go func() {
		log.Println("pprof listening:", addr)
		if err := http.ListenAndServe(addr, serverMux); err != nil {
			log.Panicln("pprof listening:", err)
		}
	}()
}
