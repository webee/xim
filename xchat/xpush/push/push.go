package push

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"xim/xchat/xpush/userinfo"
	xg "xim/xchat/xpush/xg"
	"xim/xchat/xpush/xinge"
)

// xingeConfig xinge access id and secret key
type xingeConfig struct {
	accessID  string
	secretKey string
}

// xinge config
var (
	AndroidProd = &xingeConfig{
		"2100108857",                       //正式版安卓的accessId
		"bf5b940421c7edbf620622c4d7255a12", //正式版安卓的secretKey
	}
	IosProd = &xingeConfig{
		"2200108858",                       //正式版ios的accseeId
		"87745f87793d4070fa52cfbabb0baa61", //正式版ios的secretKey
	}

	IosEnt = &xingeConfig{
		"2200141892",
		"c9991648a315bfc80bb743f021d02d12",
	}

	AndroidTest = &xingeConfig{
		"2100118679",                       //测试版安卓的accessId
		"d4bd574e15cfed0cf55b629a72072ce2", //测试版安卓的secretKey
	}
	IosTest = &xingeConfig{
		"2200118680",                       //测试版ios的accessId
		"6332c9644c963fb5d805103827f73fdf", //测试版ios的secretKey
	}

	androidActivity = "com.engdd.familytree" // android activity 用于唤醒Android
	androidClient   *xinge.Client
	iosClient       *xg.Client
)

// NewPushClient new the xinge client.
func NewPushClient(android, ios *xingeConfig) {
	accessID, err := strconv.Atoi(android.accessID)
	if err != nil {
		l.Alert("android Wrong accessID. %s", err.Error())
	}
	androidClient = xinge.NewClient(accessID, android.secretKey)
	iosClient = xg.NewClient(ios.accessID, 300, "", ios.secretKey)
}

// OfflineMsg push messages to offfline user.
func OfflineMsg(from, to, source, token, content string, chatID, interval int64,
	env byte) error {
	l.Info("PushOfflineMsg %s", to)
	ts, ok := userinfo.CheckLastPushTime(to, interval)
	if !ok {
		l.Info("PushOfflineMsg too frequently messages, so ignore some. %s %d", to, ts)
		return nil
	}
	// use userName as title
	userName, err := userinfo.GetUserName(from)
	if err != nil {
		l.Warning("GetUserName failed. %s", err.Error())
		userName = from // 名字不显示
	}
	l.Info("#user_name# %s %s", from, userName)

	var resp xinge.Response
	szChatID := fmt.Sprintf("%d", chatID)
	if strings.ToLower(source) != "appstore" {
		msg := xinge.DefaultMessage(userName, content)
		msg.Style.Clearable = 1
		msg.Style.NId = int(time.Now().Unix())
		msg.Action.ActionType = 1
		msg.Action.Activity = androidActivity
		msg.Custom = map[string]string{"chat_id": szChatID}

		resp = androidClient.PushSingleDevice(xinge.Android, token, msg)
		if resp.Code != 0 {
			return errors.New(resp.Msg)
		}
		return nil
	}
	aps := &xg.ApsAttr{Alert: userName + "发来一条消息", Badge: 1, Sound: "bingbong.aiff"}
	message := &xg.IosMessage{Aps: aps, CustomContent: map[string]interface{}{"chat_id": szChatID}}
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
