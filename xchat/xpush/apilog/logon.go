package apilog

import (
	"encoding/json"
	"net/http"
	"net/url"
	"xim/xchat/xpush/userinfo"
)

const (
	API_LOG_ONLINE  = "/apilog/usr/online"
	API_LOG_OFFLINE = "/apilog/usr/offline"
)

var (
	ApiLogHost = "http://apilogdoc.engdd.com"
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

func InitApiLogHost(host string) {
	ApiLogHost = host
}

func ApiLog(uri, userId, source string, params map[string]interface{}) error {
	client := &http.Client{}
	v := url.Values{}
	v.Add("uid", userId)
	v.Add("source", source)
	ret, err := json.Marshal(params)
	if err != nil {
		l.Error("json.Marshal failed. %v", err)
		v.Add("params", "")
	} else {
		v.Add("params", string(ret))
	}
	signed := userinfo.SecuritySuffix(uri+"?"+v.Encode())
	req, err := http.NewRequest("POST", ApiLogHost+signed, nil)
	if err != nil {
		l.Error("http.NewRequest failed. %v", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "swagger")

	resp, err := client.Do(req)
	if err != nil {
		l.Error("client.Do failed. %v", err)
		return err
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var result map[string]interface{}
	decoder.Decode(&result)

	l.Debug("%v", result)

	return nil
}

func LogOnLine(userID, source string, params map[string]interface{}) error {
	return ApiLog(API_LOG_ONLINE, userID, source, params)
}

func LogOffLine(userID, source string, params map[string]interface{}) error {
	return ApiLog(API_LOG_OFFLINE, userID, source, params)
}

