package kafka

import (
	"encoding/json"
)

// device_token，source, device_id, os_version, device_model
type UserDeviceInfo struct {
	DeviceToken string `json:"device_token"`
	Source      string `json:"source"`
	DeviceId    string `json:"device_id"`
	OsVersion   string `json:"os_version"`
	DeviceModel string `json:"device_model"`
	Update      int64  `json:"update"`
}

//
//type OnLineInfo struct {
//	User    string         `json:"user"`
//	DevInfo UserDeviceInfo `json:"info"`
//}
//
//type OffLineInfo struct {
//	User string `json:user`
//}

type LogInfo struct {
	Type string `json:"type"`
	User string `json:"user"`
	Info string `json:"info,omitempty"`
}

type MsgInfo struct {
	User     string `json:"user"`
	ChatId   int64  `json:"chat_id"`
	ChatType string `json:"chat_type"`
	From     string `json:"from"`
	Msg      string `json:"msg"`
}

//
//func Marshal(oli *OnLineInfo) ([]byte, error) {
//	ret, err := json.Marshal(oli)
//	if err != nil {
//		log.Println("json.Marshal OnLineInfo failed.", err)
//		return nil, err
//	}
//
//	return ret, nil
//}

func UnmarshalLogInfo(data []byte) (*LogInfo, error) {
	li := &LogInfo{}
	err := json.Unmarshal(data, li)
	return li, err
}

//
//func UnmarshalOffLineInfo(data []byte, oli *LogInfo) error {
//	return json.Unmarshal(data, oli)
//}

func UnmarshalMsgInfo(data []byte) (*MsgInfo, error) {
	mi := &MsgInfo{}
	err := json.Unmarshal(data, mi)
	return mi, err
}
