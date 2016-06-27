package main

import (
	"flag"
	"runtime"
	"xim/utils/pprofutils"
	"xim/xchat/logic/logger"

	"encoding/json"
	"strings"
	"time"
	"xim/xchat/xpush/apilog"
	"xim/xchat/xpush/db"
	"xim/xchat/xpush/mq"
	"xim/xchat/xpush/push"
	"xim/xchat/xpush/userinfo"
)

var (
	l = logger.Logger
)

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())
	if !args.debug {
		l.MaxLevel = 6
	}
	defer l.Close()

	if args.debug {
		pprofutils.StartPProfListen(args.pprofAddr)
	}
	setupKeys()

	apilog.InitApiLogHost(args.apiLogHost)
	userinfo.InitUserInfoHost(args.userInfoHost)

	defer db.InitRedisPool(args.redisAddr, args.redisPassword, args.poolSize)()
	push.NewPushClient(args.xgtest)

	ConsumeMsg()
	ConsumeLog()

	setupSignal()
}

// 消费消息
func ConsumeMsg() {
	consumeMsgChan := make(chan []byte, 1024)
	go mq.ConsumeGroup(args.zkAddr, mq.CONSUME_MSG_GROUP, mq.XCHAT_MSG_TOPIC, 0, 0, consumeMsgChan)
	go func() {
		for {
			select {
			case data := <-consumeMsgChan:
				l.Info("###consumeMsg###%s", string(data))
				msg, err := mq.UnmarshalMsgInfo(data)
				if err != nil {
					l.Error("mq.UnmarshalMsgInfo failed. %v", err)
					continue
				}
				udi, err := db.GetUserDeviceInfo(msg.User)
				if err != nil {
					l.Error("GetUserDeviceInfo failed. %v", err)
				} else {
					var env byte
					if args.xgtest {
						env = 2
					} else {
						env = 1
					}
					timestamp, err := time.Parse(time.RFC3339Nano, msg.Ts)
					if err != nil {
						l.Warning("time.Parse failed. %s", err.Error())
						continue
					}
					if timestamp.Unix()+int64(60) > time.Now().Unix() {
						err = push.PushOfflineMsg(msg.From, msg.User, udi.Source, udi.DeviceToken, msg.ChatId, args.pushInterval, env)
						if err != nil {
							l.Error("push.PushOfflineMsg failed. %v", err)
						}
					}
				}
			}
		}
	}()
}

// 消费登录日志
func ConsumeLog() {
	consumeLogChan := make(chan []byte, 1024)
	go mq.ConsumeGroup(args.zkAddr, mq.CONSUME_LOG_GROUP, mq.XCHAT_LOG_TOPIC, 0, 0, consumeLogChan)
	go func() {
		for {
			select {
			case data := <-consumeLogChan:
				l.Info("###consumeLog### %s", string(data))
				msg, err := mq.UnmarshalLogInfo(data)
				if err != nil {
					l.Error("mq.UnmarshalLogInfo failed. %v", err)
					continue
				}
				var udi mq.UserDeviceInfo
				err = json.Unmarshal([]byte(msg.Info), &udi)
				if err != nil {
					l.Error("Error: json.Unmarshal failed. %v", err)
					continue
				}

				params := make(map[string]interface{}, 8)
				params["device_token"] = udi.DeviceToken
				params["device_id"] = udi.DeviceId
				params["os_version"] = udi.OsVersion
				params["device_model"] = udi.DeviceModel

				logType := strings.ToLower(msg.Type)
				if "online" == logType {
					// 设置token
					udi.Update = time.Now().Unix()
					err = db.SetUserDeviceInfo(msg.User, &udi)
					if err != nil {
						l.Error("Error: SetUserDeviceInfo failed. %v", err)
					}
					err = apilog.LogOnLine(msg.User, udi.Source, params)
					if err != nil {
						l.Error("LogOnLine failed. %v", err)
					} else {
						l.Debug("LogOnLine success.")
					}
				} else if "offline" == logType {
					err = apilog.LogOffLine(msg.User, udi.Source, params)
					if err != nil {
						l.Error("LogOffLine failed. %v", err)
					} else {
						l.Debug("LogOffLine success.")
					}
				} else {
					l.Error("Error: Unknown log type")
				}
			}
		}
	}()
}
