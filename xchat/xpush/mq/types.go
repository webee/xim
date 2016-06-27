package mq

import (
	"encoding/json"
)

// UserDeviceInfo device_token，source, device_id, os_version, device_model
type UserDeviceInfo struct {
	DeviceToken string `json:"device_token"`
	Source      string `json:"source"`
	DeviceID    string `json:"device_id"`
	OsVersion   string `json:"os_version"`
	DeviceModel string `json:"device_model"`
	Update      int64  `json:"update"`
}

// LogInfo log info struct
type LogInfo struct {
	Type string `json:"type"`
	User string `json:"user"`
	Info string `json:"info,omitempty"`
}

// MsgInfo message info struct
type MsgInfo struct {
	User     string `json:"user"`
	ChatID   int64  `json:"chat_id"`
	ChatType string `json:"chat_type"`
	From     string `json:"from"`
	Msg      string `json:"msg"`
	Ts       string `json:"ts"`
}

// UnmarshalLogInfo unmarshal user log info
func UnmarshalLogInfo(data []byte) (*LogInfo, error) {
	li := &LogInfo{}
	err := json.Unmarshal(data, li)
	return li, err
}

// UnmarshalMsgInfo unmarshal user message info
func UnmarshalMsgInfo(data []byte) (*MsgInfo, error) {
	mi := &MsgInfo{}
	err := json.Unmarshal(data, mi)
	return mi, err
}
