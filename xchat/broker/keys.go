package main

import (
	"io/ioutil"
	"log"
)

var (
	userKey   []byte
	csUserKey []byte
)

func setupKeys() {
	var err error
	// user key
	userKeyPath := args.userKeyPath

	if userKey, err = ioutil.ReadFile(userKeyPath); err != nil {
		log.Fatalln(err)
	}

	// cs user key
	csUserKeyPath := args.userKeyPath

	if csUserKey, err = ioutil.ReadFile(csUserKeyPath); err != nil {
		log.Fatalln(err)
	}

	if args.debug {
		log.Println("userKey: ", string(userKey))
		log.Println("csUserKey: ", string(csUserKey))
	}
}
