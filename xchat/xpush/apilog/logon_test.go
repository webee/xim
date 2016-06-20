package apilog

import "testing"

func TestLogOnLine(t *testing.T) {
	err := LogOnLine("88888888", "google", nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLogOffLine(t *testing.T) {
	err := LogOffLine("77482", "google", nil)
	if err != nil {
		t.Fatal(err)
	}
}
