package userinfo

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

func GetSign(url string, deviceId string, time string) string {
	var buffer bytes.Buffer
	buffer.WriteString(url)
	buffer.WriteString(deviceId)
	buffer.WriteString(time)

	md5Str := md5Encrypt(buffer.Bytes())
	md5Str2 := mixup([]byte(md5Str))
	fixSalt := "d32dd3ba42e32f53f4ff7f3f6a6c92ce"

	buffer.Reset()
	buffer.WriteString(fixSalt)
	buffer.Write(md5Str2)
	return md5Encrypt(buffer.Bytes())
}

func md5Encrypt(data []byte) (res string) {
	h := md5.New()
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

/**
混淆
*/
func mixup(data []byte) (res []byte) {
	exchange(data, 1, 6)
	exchange(data, 2, 12)
	exchange(data, 14, 22)
	return data
}

func exchange(data []byte, x int, y int) {
	tmp := data[x]
	data[x] = data[y]
	data[y] = tmp
}

func SecuritySuffix(url string) string {
	deviceId := "1234567890123456789012345678901234567890"
	timeStamp := time.Now().UTC().Format("20060102150405")
	//timeStamp:=time.Now().UTC().Format("2006-01-02 15:04:05")
	if strings.HasSuffix(url, "?") {
		url += "&timestamp=" + timeStamp
	} else {
		if strings.HasSuffix(url, "&") {
			url += "timestamp=" + timeStamp
		} else {
			url += "&timestamp=" + timeStamp
		}

	}
	url += "&deviceId=" + deviceId
	sign := GetSign(url, deviceId, timeStamp)
	url += "&sign=" + sign
	return url
}

type UserInfo struct {
	User   string
	Name   string
	update int64
}

type PushInterval struct {
	ts int64
}

const (
	USER_NAME_CACHE_VALID_PERIOD = 600
	OFFLINE_MSG_PUSH_INTERVAL    = 60 // Second
)

var (
	UserInfoCache    = make(map[string]*UserInfo, 10000)
	URL              = "http://test.engdd.com"
	Produce_URL      = ""
	UserPushInterval = make(map[string]PushInterval, 10000)
)

func CheckLastPushTime(user string) (int64, bool) {
	now := time.Now().Unix()
	ts, ok := UserPushInterval[user]
	if ok && now-ts.ts < OFFLINE_MSG_PUSH_INTERVAL {
		return ts.ts, false
	} else {
		UserPushInterval[user] = PushInterval{now} // 更新最后发送时间
		return ts.ts, true
	}
}

func FetchUserName(uid string) (string, error) {
	uri := SecuritySuffix(fmt.Sprintf("/login/findOtherUser/1/%s?", uid))
	urlStr := fmt.Sprintf("%s%s", URL, uri)

	client := http.Client{}
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		log.Println("http.NewRequest failed.", err)
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Println("client.Do failed.", err)
		return "", err
	}

	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var ret map[string]interface{}
	err = decoder.Decode(&ret)
	if err != nil {
		log.Println("json.Decode failed.", err)
		return "", err
	}

	obj, ok := ret["obj"]
	if ok {
		objMap, ok := obj.(map[string]interface{})
		person, ok := objMap["person"]
		personMap, ok := person.(map[string]interface{})
		if ok {
			fullName, ok := personMap["fullName"]
			if ok {
				name, ok := fullName.(string)
				if ok {
					log.Println("#fetch_user_name", uid, name)
					return name, nil
				}
			}
		}
	}

	return "", errors.New("user not found")
}

func GetUserName(uid string) (string, error) {
	ui, ok := UserInfoCache[uid]
	if ok {
		// 检查缓存是否过期
		if ui.update+USER_NAME_CACHE_VALID_PERIOD > time.Now().Unix() {
			log.Println("hit the cache", uid, ui.Name)
			return ui.Name, nil
		}
	}

	fullName, err := FetchUserName(uid)
	if err != nil {
		log.Println("FetchUserName failed.", err)
		return "", err
	}
	// 设置缓存，后续可改为异步设置
	UserInfoCache[uid] = &UserInfo{uid, fullName, time.Now().Unix()}
	return fullName, nil
}
