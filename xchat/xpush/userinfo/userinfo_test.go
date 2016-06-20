package userinfo

import (
	"testing"
	"log"
)


func TestGetUserName(t *testing.T) {
	name, err := GetUserName("77482")
	if err != nil {
		t.Fatal(err)
	}
	log.Println(name)
}

func BenchmarkGetUserName(b *testing.B) {

}
