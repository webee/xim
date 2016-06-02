package main

import (
	"io/ioutil"
	"log"
)

var (
	userKey []byte
)

func setupKeys() {
	var err error
	userKeyPath := args.userKeyPath

	if userKey, err = ioutil.ReadFile(userKeyPath); err != nil {
		log.Fatalln(err)
	}

	if args.debug {
		log.Println("userKey: ", string(userKey))
	}
}
