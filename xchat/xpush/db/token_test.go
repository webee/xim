package db

import (
	"testing"
	"time"
	"xim/xchat/xpush/mq"
)

func TestSetUserDeviceInfo(t *testing.T) {
	udi := &mq.UserDeviceInfo{DeviceToken: "412936f4d21e80d84a77a7c756bd03e2da2f1c2e",
		Source: "google", DeviceID: "device_id", OsVersion: "os_version",
		DeviceModel: "android", Update: time.Now().Unix()}
	err := SetUserDeviceInfo("77482", udi)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetUserDeviceInfo(t *testing.T) {
	udi, err := GetUserDeviceInfo("77482")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(udi)
}
