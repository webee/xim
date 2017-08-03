package main

import (
	"io/ioutil"
	"log"
	"xim/utils/commons"
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
	redisPassword = commons.TrimBytesSpace(redisPassword)

	args.redisPassword = string(redisPassword)

	if args.debug {
		log.Println("redis password: ", args.redisPassword)
	}
}
