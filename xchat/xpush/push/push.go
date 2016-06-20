//package main
package push

import (
	"github.com/LibiChai/xinge"
	"time"
	"errors"
	"log"
	"strings"
	"xim/xchat/xpush/userinfo"
)

const (
	ACCESS_ID_ANDROID  = 2100108857                         //正式版安卓的accessId
	SECRET_KEY_ANDROID = "bf5b940421c7edbf620622c4d7255a12" //正式版安卓的secretKey
	ACCESS_ID_IOS      = 2200108858                         //正式版ios的accseeId
	SECRET_KEY_IOS     = "87745f87793d4070fa52cfbabb0baa61" //正式版ios的secretKey

	ACCESS_ID_IOS_ENT  = 2200141892
	SECRET_KEY_IOS_ENT = "c9991648a315bfc80bb743f021d02d12"

	//测试版属性
	ACCESS_ID_ANDROID_TEST  = 2100118679                         //测试版安卓的accessId
	SECRET_KEY_ANDROID_TEST = "d4bd574e15cfed0cf55b629a72072ce2" //测试版安卓的secretKey
	ACCESS_ID_IOS_TEST      = 2200118680                         //测试版ios的accessId
	SECRET_KEY_IOS_TEST     = "6332c9644c963fb5d805103827f73fdf" //测试版ios的secretKey

	ANDROID_ACTIVITY = "com.engdd.familytree" // android activity 用于唤醒Android
)

var (
	androidClient = xinge.NewClient(ACCESS_ID_ANDROID_TEST, SECRET_KEY_ANDROID_TEST)
	iosClient     = xinge.NewClient(ACCESS_ID_IOS_TEST, SECRET_KEY_IOS_TEST)
)

func PushOfflineMsg(user, dev, token, chatId string) error {
	// use userName as title
	userName, err := userinfo.GetUserName(user)
	if err != nil {
		log.Println("GetUserName failed.", err)
		userName = user // 名字不显示
	}
	log.Println("#user_name#", user, userName)

	var resp xinge.Response
	if strings.ToLower(dev) == "android" {
		msg := xinge.DefaultMessage(userName, "Hello HanMeimei")
		msg.Style.Clearable = 1
		msg.Style.NId = int(time.Now().Unix())
		msg.Action.ActionType = 1
		msg.Action.Activity = ANDROID_ACTIVITY
		msg.Custom = map[string]string{"chat_id": chatId}

		resp = androidClient.PushSingleDevice(xinge.Android, token, msg)
	} else if strings.ToLower(dev) == "ios" {
		resp = iosClient.PushSingleIosDevice(token, userName, 5, map[string]string{"hello": "hello there"})
	}

	if resp.Code != 0 {
		return errors.New(resp.Msg)
	}

	return nil
}

//
//func PushGroupMessage() {
//	client := xinge.NewClient(ACCESS_ID_ANDROID_TEST, SECRET_KEY_ANDROID_TEST)
//	msg := xinge.DefaultMessage("测试群", "LiLei: "+"Hello There!")
//	client.PushSingleDevice(xinge.Android, "412936f4d21e80d84a77a7c756bd03e2da2f1c2e", msg)
//}
