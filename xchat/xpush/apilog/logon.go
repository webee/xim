package apilog

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"github.com/go-mangos/mangos/protocol/req"
)

const (
	API_LOG_HOST    = "http://apilogdoc.engdd.com"
	API_LOG_ONLINE  = "/apilog/usr/online"
	API_LOG_OFFLINE = "/apilog/usr/offline"
)

type Log struct {
	Uid    int64  `json:"uid"`
	Source string `json:"source"`
	Params string `json:"params"`
}

type OffLine struct {
	User int
	Info map[string]interface{}
}

func ApiLog(uri, userId, source string, params map[string]interface{}) error {
	client := &http.Client{}
	v := url.Values{}
	v.Add("uid", userId)
	v.Add("source", source)
	v.Add("params", json.Marshal(params))
	req, err := http.NewRequest("POST", API_LOG_HOST+uri+"?"+v.Encode(), nil)
	if err != nil {
		log.Println("http.NewRequest failed.", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "swagger")

	resp, err := client.Do(req)
	if err != nil {
		log.Println("client.Do failed.", err)
		return err
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var result map[string]interface{}
	decoder.Decode(&result)

	log.Println(result)

	return nil
}

func LogOnLine(userID, source string, params map[string]interface{}) error {
	return ApiLog(API_LOG_ONLINE, userID, source, params)
}

func LogOffLine(userID, source string, params map[string]interface{}) error {
	return ApiLog(API_LOG_OFFLINE, userID, source, params)
}
//
//func main() {
//	LogOnLine("88888888", "google", nil)
//}
