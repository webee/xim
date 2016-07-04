package push

import "testing"

func TestPushOfflineMsg(t *testing.T) {
	// 测试android
	NewPushClient(AndroidProd, IosProd)
	err := OfflineMsg("82312", "77481", "Android", "412936f4d21e80d84a77a7c756bd03e2da2f1c2e", "Hello", 123456, 30, 2)
	if err != nil {
		t.Fatal(err)
	}

}
