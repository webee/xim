package userboard

import (
	"io/ioutil"
	"log"
)

var (
	appKey  []byte
	userKey []byte
)

func setupKeys(config *Config) {
	var err error
	appKeyPath := config.AppKeyPath
	if appKey, err = ioutil.ReadFile(appKeyPath); err != nil {
		log.Fatalln(err)
	}

	userKeyPath := config.UserKeyPath
	if userKey, err = ioutil.ReadFile(userKeyPath); err != nil {
		log.Fatalln(err)
	}

	if config.Debug {
		log.Println("appKey: ", string(appKey))
		log.Println("userKey: ", string(userKey))
	}
}
