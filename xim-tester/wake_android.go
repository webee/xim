package main

import (
	"fmt"
	"github.com/aiwuTech/xinge"
	"time"
)

var (
	xingeClient = xinge.NewClient("2100118679", 300, "2100118679", "d4bd574e15cfed0cf55b629a72072ce2")
)

func main() {
	pack := &xinge.Package{PackageName: "com.engdd.familytree", Confirm: 1}
	action := &xinge.AndroidAction{ActionType: 4, PackageName: pack}
	szChatID := "12345"
	message := &xinge.AndroidMessage{
		Content:       "wifi密码qq发过来",
		Title:         "wifi密码是多少",
		Action:        action,
		CustomContent: map[string]interface{}{"chat_id": szChatID},
	}

	reqPush := &xinge.ReqPush{
		PushType: xinge.PushType_single_device,
		DeviceToken:  "412936f4d21e80d84a77a7c756bd03e2da2f1c2e",
		MessageType:  xinge.MessageType_notify,
		Message:      message,
		ExpireTime:   300,
		SendTime:     time.Now(),
		MultiPkgType: xinge.MultiPkg_aid,
		PushEnv:      xinge.PushEnv_android,
		PlatformType: xinge.Platform_android,
		LoopTimes:    2,
		LoopInterval: 7,
		Cli:          xingeClient,
	}
	fmt.Println(reqPush.Push())

	fmt.Println(xingeClient.AppDeviceNum())
}
