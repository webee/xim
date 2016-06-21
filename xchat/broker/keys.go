package main

import (
	"io/ioutil"
	"log"
)

var (
	userKey     []byte
	testUserKey []byte
	csUserKey   []byte
)

func setupKeys() {
	var err error
	// user key
	userKeyPath := args.userKeyPath

	if userKey, err = ioutil.ReadFile(userKeyPath); err != nil {
		log.Fatalln(err)
	}

	// test user key
	testUserKeyPath := args.testUserKeyPath

	if testUserKey, err = ioutil.ReadFile(testUserKeyPath); err != nil {
		log.Fatalln(err)
	}

	// cs user key
	csUserKeyPath := args.csUserKeyPath

	if csUserKey, err = ioutil.ReadFile(csUserKeyPath); err != nil {
		log.Fatalln(err)
	}

	if args.debug {
		log.Println("userKey: ", string(userKey))
		log.Println("testUserKey: ", string(testUserKey))
		log.Println("csUserKey: ", string(csUserKey))
	}
}
