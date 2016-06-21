package main

import (
	"flag"
	"log"
	"runtime"
	"xim/utils/pprofutils"
	"xim/xchat/logic/logger"

	"encoding/json"
	"strings"
	"xim/xchat/xpush/apilog"
	"xim/xchat/xpush/kafka"
	"xim/xchat/xpush/push"
	"xim/xchat/xpush/token"
	"time"
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

	ConsumeMsg()
	ConsumeLog()

	setupSignal()
}

// 消费消息
func ConsumeMsg() {
	consumeMsgChan := make(chan []byte, 1024)
	go kafka.ConsumeGroup(args.zkAddr, kafka.CONSUME_MSG_GROUP, kafka.XCHAT_MSG_TOPIC, 0, 0, consumeMsgChan)
	go func() {
		for {
			select {
			case data := <-consumeMsgChan:
				log.Println("###consumeMsg###", string(data))
				msg, err := kafka.UnmarshalMsgInfo(data)
				if err != nil {
					log.Println("kafka.UnmarshalMsgInfo failed.", err)
					continue
				}
				udi, err := token.GetUserDeviceInfo(args.redisAddr, msg.User)
				if err != nil {
					log.Println("GetUserDeviceInfo failed.", err)
				} else {
					err = push.PushOfflineMsg(msg.From, udi.DeviceModel, udi.DeviceToken, msg.ChatId)
					if err != nil {
						log.Println("push.PushOfflineMsg failed.", err)
					}
				}
			}
		}
	}()
}

// 消费登录日志
func ConsumeLog() {
	consumeLogChan := make(chan []byte, 1024)
	go kafka.ConsumeGroup(args.zkAddr, kafka.CONSUME_LOG_GROUP, kafka.XCHAT_LOG_TOPIC, 0, 0, consumeLogChan)
	go func() {
		for {
			select {
			case data := <-consumeLogChan:
				log.Println("###consumeLog###", string(data))
				msg, err := kafka.UnmarshalLogInfo(data)
				if err != nil {
					log.Println("kafka.UnmarshalLogInfo failed.", err)
					continue
				}
				var udi kafka.UserDeviceInfo
				err = json.Unmarshal([]byte(msg.Info), &udi)
				if err != nil {
					log.Println("Error: json.Unmarshal failed.", err)
					continue
				}
				udi.Update = time.Now().Unix()
				err = token.SetUserDeviceInfo(args.redisAddr, msg.User, &udi)
				if err != nil {
					log.Println("Error: SetUserDeviceInfo failed.", err)
				}

				logType := strings.ToLower(msg.Type)
				if "online" == logType {
					err = apilog.LogOnLine(msg.User, udi.Source, map[string]interface{}{"param": msg.Info})
					if err != nil {
						log.Println("LogOnLine failed.", err)
					} else {
						log.Println("LogOnLine success.")
					}
				} else if "offline" == logType {
					err = apilog.LogOffLine(msg.User, udi.Source, map[string]interface{}{"param": msg.Info})
					if err != nil {
						log.Println("LogOffLine failed.", err)
					} else {
						log.Println("LogOffLine success.")
					}
				} else {
					log.Println("Error: Unknown log type")
				}
			}
		}
	}()
}
