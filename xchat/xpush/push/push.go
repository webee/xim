package push

import (
	"errors"
	"fmt"
	"strings"
	"time"
	xg "xim/xchat/xpush/push/xg"
	"xim/xchat/xpush/userinfo"
	"xim/xchat/xpush/xinge"
)

const (
	ACCESS_ID_ANDROID  = 2100108857                         //正式版安卓的accessId
	SECRET_KEY_ANDROID = "bf5b940421c7edbf620622c4d7255a12" //正式版安卓的secretKey
	ACCESS_ID_IOS      = "2200108858"                       //正式版ios的accseeId
	SECRET_KEY_IOS     = "87745f87793d4070fa52cfbabb0baa61" //正式版ios的secretKey

	ACCESS_ID_IOS_ENT  = 2200141892
	SECRET_KEY_IOS_ENT = "c9991648a315bfc80bb743f021d02d12"

	//测试版属性
	ACCESS_ID_ANDROID_TEST  = 2100118679                         //测试版安卓的accessId
	SECRET_KEY_ANDROID_TEST = "d4bd574e15cfed0cf55b629a72072ce2" //测试版安卓的secretKey
	ACCESS_ID_IOS_TEST      = "2200118680"                       //测试版ios的accessId
	SECRET_KEY_IOS_TEST     = "6332c9644c963fb5d805103827f73fdf" //测试版ios的secretKey

	ANDROID_ACTIVITY = "com.engdd.familytree" // android activity 用于唤醒Android
)

var (
	androidClient *xinge.Client
	iosClient     *xg.Client
)

func NewPushClient(testing bool) {
	if testing {
		androidClient = xinge.NewClient(ACCESS_ID_ANDROID_TEST, SECRET_KEY_ANDROID_TEST)
		iosClient = xg.NewClient(ACCESS_ID_IOS_TEST, 300, "", SECRET_KEY_IOS_TEST)
	} else {
		androidClient = xinge.NewClient(ACCESS_ID_ANDROID, SECRET_KEY_ANDROID)
		//iosClient = xinge.NewClient(ACCESS_ID_IOS, SECRET_KEY_IOS)
		iosClient = xg.NewClient(ACCESS_ID_IOS, 300, "", SECRET_KEY_IOS)
	}
}

func PushOfflineMsg(from, to, source, token string, chatId, interval int64, env byte) error {
	l.Info("PushOfflineMsg %s", to)
	ts, ok := userinfo.CheckLastPushTime(to, interval)
	if !ok {
		l.Info("PushOfflineMsg too frequently messages, so ignore some. %s %d", to, ts)
		return nil
	}
	// use userName as title
	userName, err := userinfo.GetUserName(from)
	if err != nil {
		l.Error("GetUserName failed. %v", err)
		userName = from // 名字不显示
	}
	l.Info("#user_name# %s %s", from, userName)

	var resp xinge.Response
	szChatId := fmt.Sprintf("%d", chatId)
	if strings.ToLower(source) != "appstore" {
		msg := xinge.DefaultMessage(userName, "发来一条消息")
		msg.Style.Clearable = 1
		msg.Style.NId = int(time.Now().Unix())
		msg.Action.ActionType = 1
		msg.Action.Activity = ANDROID_ACTIVITY
		msg.Custom = map[string]string{"chat_id": szChatId}

		resp = androidClient.PushSingleDevice(xinge.Android, token, msg)
		if resp.Code != 0 {
			return errors.New(resp.Msg)
		}
		return nil
	} else {
		aps := &xg.ApsAttr{Alert: userName + "发来一条消息", Badge: 1, Sound: "bingbong.aiff"}
		message := &xg.IosMessage{aps, map[string]interface{}{"chat_id": szChatId}}
		reqPush := &xg.ReqPush{
			PushType:     xg.PushType_single_device,
			TagsOp:       xg.TagsOp_AND,
			DeviceToken:  token,
			MessageType:  xg.MessageType_ios,
			Message:      message,
			ExpireTime:   0,
			SendTime:     time.Now(),
			MultiPkgType: xg.MultiPkg_ios,
			PushEnv:      xg.PushEnv(env),
			PlatformType: xg.Platform_ios,
			LoopTimes:    2,
			LoopInterval: 7,
			Cli:          iosClient,
		}
		return reqPush.Push()
	}
}
