package main

import (
	"io/ioutil"
	"log"
)

var (
	redisPassword []byte
)

func setupKeys() {
	var err error
	// user key
	redisPasswordPath := args.redisPasswordPath

	if redisPassword, err = ioutil.ReadFile(redisPasswordPath); err != nil {
		log.Fatalln(err)
	}

	args.redisPassword = string(redisPassword)

	if args.debug {
		log.Println("redis password: ", args.redisPassword)
	}
}
