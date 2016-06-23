package db

import (
)

type UserDeviceInfo struct {
	user int `json:"user"`
	kafka.UserDeviceInfo
}
//
//type KafkaOffset struct {
//	Topic  string `json:"topic"`
//	Offset int    `json:"offset"`
//}
