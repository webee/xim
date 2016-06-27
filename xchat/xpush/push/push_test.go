package push

import "testing"

func TestPushOfflineMsg(t *testing.T) {
	// 测试android
	err := OfflineMsg("77482", "77481", "Android", "412936f4d21e80d84a77a7c756bd03e2da2f1c2e", 123456, 30, 2)
	if err != nil {
		t.Fatal(err)
	}

	// 测试ios
	//err := PushOfflineMsg("")
}

//client := xinge.NewClient(ACCESS_ID_ANDROID_TEST, SECRET_KEY_ANDROID_TEST)
//client.PushSingleAndroidDevice("412936f4d21e80d84a77a7c756bd03e2da2f1c2e", "hello", "boooom", nil)
//
//func PushSingleDevice(title, content, deviceToken string) error {
//	message := &xinge.AndroidMessage{
//		Content: content,
//		Title:   title,
//		//AcceptTime: []*xinge.AcceptTime{&xinge.AcceptTime{Start:&xinge.HourMin{"9","0"}, End:&xinge.HourMin{"21", "0"}}},
//	}
//
//	reqPush := &xinge.ReqPush{
//		PushType:     xinge.PushType_single_device,
//		TagsOp:       xinge.TagsOp_AND,
//		DeviceToken:  deviceToken,
//		MessageType:  xinge.MessageType_notify,
//		Message:      message,
//		ExpireTime:   300,
//		SendTime:     time.Now(),
//		MultiPkgType: xinge.MultiPkg_aid,
//		PushEnv:      xinge.PushEnv_android,
//		PlatformType: xinge.Platform_android,
//		LoopTimes:    2,
//		LoopInterval: 7,
//		Cli:          xingeClient,
//	}
//
//	return reqPush.Push()
//}

//func main() {
//	//PushSingleDevice("hello", "bomb, click", "412936f4d21e80d84a77a7c756bd03e2da2f1c2e")
//	client := xinge.NewClient(ACCESS_ID_ANDROID_TEST, SECRET_KEY_ANDROID_TEST)
//	log.Println("###PUSH Android###")
//	//client.PushSingleAndroidDevice("412936f4d21e80d84a77a7c756bd03e2da2f1c2e", "hello", "boooom", nil)
//
//	msg := xinge.DefaultMessage("LiLei", "Hello HanMeimei")
//	msg.Style.Clearable = 1
//	msg.Style.NId = int(time.Now().Unix())
//	msg.Action.ActionType = 1
//	msg.Action.Activity = "com.engdd.familytree"
//	msg.Custom = map[string]string{"chat_id":"123456", "sender":"Tom", "msg":""}
//	client.PushSingleDevice(xinge.Android, "412936f4d21e80d84a77a7c756bd03e2da2f1c2e", msg)
//
//	msg = xinge.DefaultMessage("测试群", "LiLei: " + "Hello There!" )
//	client.PushSingleDevice(xinge.Android, "412936f4d21e80d84a77a7c756bd03e2da2f1c2e", msg)
//
//	iClient := xinge.NewClient(ACCESS_ID_IOS_TEST, SECRET_KEY_IOS_TEST)
//	log.Println("###PUSH IOS###")
//	iClient.PushSingleIosDevice("3049a2ab83e49066ee079f6d071fc04a27b9ae0b5380c9a13ec876ae96f07a6f",
//		"boooom", 5, map[string]string{"hello": "hello there"})
//}
//
