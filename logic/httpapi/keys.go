package httpapi

import (
	"io/ioutil"
	"log"
)

var (
	salt    []byte
	appKey  []byte
	userKey []byte
)

func setupKeys(config *ServerConfig) {
	var err error
	appKeyPath := config.AppKeyPath
	userKeyPath := config.UserKeyPath

	if appKey, err = ioutil.ReadFile(appKeyPath); err != nil {
		log.Fatalln(err)
	}

	if userKey, err = ioutil.ReadFile(userKeyPath); err != nil {
		log.Fatalln(err)
	}

	if config.Debug {
		log.Println("salt: ", string(salt))
		log.Println("appKey: ", string(appKey))
		log.Println("userKey: ", string(userKey))
	}
}
