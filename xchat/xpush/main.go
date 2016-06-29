package main

import (
	"flag"
	"fmt"
	"runtime"
	"xim/utils/pprofutils"
	"xim/xchat/logic/logger"

	"encoding/json"
	"strings"
	"time"
	"xim/xchat/xpush/apilog"
	"xim/xchat/xpush/db"
	"xim/xchat/xpush/immsg"
	"xim/xchat/xpush/mq"
	"xim/xchat/xpush/push"
	"xim/xchat/xpush/userinfo"
)

var (
	l = logger.Logger
)

func main() {
	flag.Parse()
	fmt.Println("args", args)
	runtime.GOMAXPROCS(runtime.NumCPU())
	if !args.debug {
		l.MaxLevel = 6
	}
	defer l.Close()

	if args.debug {
		pprofutils.StartPProfListen(args.pprofAddr)
	}
	setupKeys()

	apilog.InitAPILogHost(args.apiLogHost)
	userinfo.InitUserInfoHost(args.userInfoHost)

	defer db.InitRedisPool(args.redisAddr, args.redisPassword, args.poolSize)()
	if args.xgtest {
		l.Info("testing push")
		push.NewPushClient(push.AndroidTest, push.IosTest)
	} else {
		l.Info("product push")
		push.NewPushClient(push.AndroidProd, push.IosProd)
	}
	consumeMsg()
	consumeLog()

	setupSignal()
}

// 消费消息
func consumeMsg() {
	var env byte
	if args.xgtest {
		env = 2
	} else {
		env = 1
	}
	consumeMsgChan := make(chan []byte, 1024)
	go mq.ConsumeGroup(args.zkAddr, mq.ConsumeMsgGroup, mq.XchatMsgTopic, 0, 0, consumeMsgChan)
	go func() {
		for {
			data := <-consumeMsgChan
			l.Info("###consumeMsg###%s", string(data))
			msg, err := mq.UnmarshalMsgInfo(data)
			if err != nil {
				l.Warning("mq.UnmarshalMsgInfo failed. %s", err.Error())
				continue
			}
			timestamp, err := time.Parse(time.RFC3339Nano, msg.Ts)
			if err != nil {
				l.Warning("time.Parse failed. %s", err.Error())
				continue
			}
			if timestamp.Unix()+int64(60) < time.Now().Unix() {
				l.Warning("message out of data. %s", msg.Ts)
				continue
			}
			content, err := immsg.ParseMsg([]byte(msg.Msg))
			if err != nil {
				l.Warning("immsg parse failed. %s", err.Error())
				continue
			}
			udi, err := db.GetUserDeviceInfo(msg.User)
			if err != nil {
				l.Warning("GetUserDeviceInfo failed. %s", err.Error())
				continue
			}
			err = push.OfflineMsg(msg.From, msg.User, udi.Source,
				udi.DeviceToken, content, msg.ChatID, args.pushInterval, env)
			if err != nil {
				l.Warning("push.PushOfflineMsg failed. %s", err.Error())
				continue
			}
		}
	}()
}

// 消费登录日志
func consumeLog() {
	consumeLogChan := make(chan []byte, 1024)
	go mq.ConsumeGroup(args.zkAddr, mq.ConsumeLogGroup, mq.XchatLogTopic, 0, 0, consumeLogChan)
	go func() {
		for {
			data := <-consumeLogChan
			l.Info("###consumeLog### %s", string(data))
			msg, err := mq.UnmarshalLogInfo(data)
			if err != nil {
				l.Warning("mq.UnmarshalLogInfo failed. %s", err.Error())
				continue
			}
			var udi mq.UserDeviceInfo
			err = json.Unmarshal([]byte(msg.Info), &udi)
			if err != nil {
				l.Warning("Error: json.Unmarshal failed. %s", err.Error())
				continue
			}

			params := make(map[string]interface{}, 8)
			params["device_token"] = udi.DeviceToken
			params["device_id"] = udi.DeviceID
			params["os_version"] = udi.OsVersion
			params["device_model"] = udi.DeviceModel

			logType := strings.ToLower(msg.Type)
			if "online" == logType {
				// 设置token
				udi.Update = time.Now().Unix()
				err = db.SetUserDeviceInfo(msg.User, &udi)
				if err != nil {
					l.Warning("Error: SetUserDeviceInfo failed. %s", err.Error())
				}
				err = apilog.LogOnLine(msg.User, udi.Source, params)
				if err != nil {
					l.Warning("LogOnLine failed. %s", err.Error())
				} else {
					l.Info("LogOnLine success.")
				}
			} else if "offline" == logType {
				err = apilog.LogOffLine(msg.User, udi.Source, params)
				if err != nil {
					l.Warning("LogOffLine failed. %s", err.Error())
				} else {
					l.Info("LogOffLine success.")
				}
			} else {
				l.Error("Error: Unknown log type")
			}
		}
	}()
}
