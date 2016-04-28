package userboard

import (
	"io/ioutil"
	"log"
)

var (
	userKey []byte
)

func setupKeys(config *Config) {
	var err error
	userKeyPath := config.UserKeyPath
	if userKey, err = ioutil.ReadFile(userKeyPath); err != nil {
		log.Fatalln(err)
	}

	if config.Debug {
		log.Println("userKey: ", string(userKey))
	}
}
