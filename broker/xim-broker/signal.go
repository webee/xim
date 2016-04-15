package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

// InitSignal register signals handler.
func InitSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	for {
		s := <-c
		log.Println("get a signal: ", s.String())
		switch s {
		case os.Interrupt, syscall.SIGTERM:
			return
		default:
			return
		}
	}
}
