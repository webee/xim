package apilog

import (
	"encoding/json"
	"net/http"
	"net/url"
	"xim/xchat/xpush/userinfo"
)

const (
	apiLogOnline  = "/apilog/usr/online"
	apiLogOffline = "/apilog/usr/offline"
)

var (
	apiLogHost = "http://apilogdoc.engdd.com"
)

// InitAPILogHost init the api log host addr
func InitAPILogHost(host string) {
	apiLogHost = host
}

func apiLog(uri, userID, source string, params map[string]interface{}) error {
	client := &http.Client{}
	v := url.Values{}
	v.Add("uid", userID)
	v.Add("source", source)
	ret, err := json.Marshal(params)
	if err != nil {
		l.Warning("get offline users error: %s", err.Error())
		v.Add("params", "")
	} else {
		v.Add("params", string(ret))
	}
	signed := userinfo.SecuritySuffix(uri + "?" + v.Encode())
	req, err := http.NewRequest("POST", apiLogHost+signed, nil)
	if err != nil {
		l.Warning("http.NewRequest failed. %s", err.Error())
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "swagger")

	resp, err := client.Do(req)
	if err != nil {
		l.Warning("client.Do failed. %s", err.Error())
		return err
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var result map[string]interface{}
	err = decoder.Decode(&result)
	if err != nil {
		l.Warning("deocde apilog response failed. %s", err.Error())
		return err
	}

	l.Info("%v", result)

	return nil
}

// LogOnLine log user logon info
func LogOnLine(userID, source string, params map[string]interface{}) error {
	return apiLog(apiLogOnline, userID, source, params)
}

// LogOffLine log user logoff info
func LogOffLine(userID, source string, params map[string]interface{}) error {
	return apiLog(apiLogOffline, userID, source, params)
}
