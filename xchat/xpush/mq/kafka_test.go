package kafka

import (
	"testing"
	"fmt"
	"time"
	//"log"
	"encoding/json"
	"log"
)

func TestSyncProduce(t *testing.T) {
	err := SyncProduce([]string{"localhost:9092"}, "test", "Hello Kafka", 0)
	if err != nil {
		t.Fatal(err)
	}
}

func TestAsyncProduce(t *testing.T) {
	ch := make(chan []byte, 1024)
	go AsyncProduce([]string{"localhost:9092"}, "test", ch)
	for i:=0; i<100; i++ {
		ch <- []byte(fmt.Sprintf("Hello Kafka %d", i))
		time.Sleep(3 * time.Millisecond)
	}
}

func TestConsume(t *testing.T) {
	ch := make(chan []byte, 1024)
	go Consume([]string{"localhost:9092"}, "test", 0, 0, ch)

	var exit bool
	for {
		select {
		case  <-ch:
			//log.Println(string(msg))
		case <-time.After(3 * time.Second):
			exit = true
			break
		}
		if exit {
			break
		}
	}
}

func TestAsyncProduce2(t *testing.T) {
	msgInfo := &MsgInfo{User:"77482", ChatId:"12345", ChatType:"user", From:"77482", Msg:"Hello"}
	ret, err := json.Marshal(msgInfo)
	if err != nil {
		t.Fatal(err)
	}
	log.Println(string(ret))
}