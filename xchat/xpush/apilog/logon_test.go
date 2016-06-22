package apilog

import "testing"

func TestLogOnLine(t *testing.T) {
	InitApiLogHost("http://apilogdoc.engdd.com")
	err := LogOnLine("88888888", "google", nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLogOffLine(t *testing.T) {
	InitApiLogHost("http://apilogdoc.engdd.com")
	err := LogOffLine("77482", "google", nil)
	if err != nil {
		t.Fatal(err)
	}
}
