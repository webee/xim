package db

import (
	"testing"
	"xim/xchat/xpush/kafka"
	"time"
)

func TestSetUserDeviceInfo(t *testing.T) {
	udi := &kafka.UserDeviceInfo{"412936f4d21e80d84a77a7c756bd03e2da2f1c2e", "google", "device_id",
		"os_version", "android", time.Now().Unix()}
	err := SetUserDeviceInfo("127.0.0.1:6379", "77482", udi)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetUserDeviceInfo(t *testing.T) {
	udi, err := GetUserDeviceInfo("localhost:6379", "77482")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(udi)
}

