package userinfo

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// GetSign compute signature.
func GetSign(url string, deviceID string, time string) string {
	var buffer bytes.Buffer
	buffer.WriteString(url)
	buffer.WriteString(deviceID)
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

// SecuritySuffix add security suffix.
func SecuritySuffix(url string) string {
	deviceID := "1234567890123456789012345678901234567890"
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
	url += "&deviceId=" + deviceID
	sign := GetSign(url, deviceID, timeStamp)
	url += "&sign=" + sign
	return url
}

// UserInfo user name cache
type UserInfo struct {
	User   string
	Name   string
	update int64
}

// PushInterval push offline message interval.
type PushInterval struct {
	ts int64
}

const (
	userNameCacheValidPeriod = 600
)

var (
	userInfoCache    = make(map[string]*UserInfo, 10000)
	url              = "http://test.engdd.com"
	userPushInterval = make(map[string]PushInterval, 10000)
)

// InitUserInfoHost init user info host
func InitUserInfoHost(host string) {
	url = host
}

// CheckLastPushTime check receiver last message time.
func CheckLastPushTime(user string, interval int64) (int64, bool) {
	now := time.Now().Unix()
	ts, ok := userPushInterval[user]
	if ok && now-ts.ts < interval {
		return ts.ts, false
	}
	userPushInterval[user] = PushInterval{now} // 更新最后发送时间
	return ts.ts, true
}

// FetchUserName fetch user name from interface.
func FetchUserName(uid string) (string, error) {
	uri := SecuritySuffix(fmt.Sprintf("/login/findOtherUser/1/%s?", uid))
	urlStr := fmt.Sprintf("%s%s", url, uri)

	client := http.Client{}
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		l.Info("http.NewRequest failed. %s", err.Error())
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		l.Warning("client.Do failed. %s", err.Error())
		return "", err
	}

	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var ret map[string]interface{}
	err = decoder.Decode(&ret)
	if err != nil {
		l.Warning("json.Decode failed. %s", err.Error())
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
					l.Info("#fetch_user_name# %s %s", uid, name)
					return name, nil
				}
			}
		}
	}

	return "", errors.New("user not found")
}

// GetUserName get user name from cache or interface.
func GetUserName(uid string) (string, error) {
	ui, ok := userInfoCache[uid]
	if ok {
		// 检查缓存是否过期
		if ui.update+userNameCacheValidPeriod > time.Now().Unix() {
			l.Debug("hit the cache, %s %s", uid, ui.Name)
			return ui.Name, nil
		}
	}

	fullName, err := FetchUserName(uid)
	if err != nil {
		l.Warning("FetchUserName failed. %s, %s", uid, err.Error())
		return "", err
	}
	// 设置缓存，后续可改为异步设置
	userInfoCache[uid] = &UserInfo{uid, fullName, time.Now().Unix()}
	return fullName, nil
}
