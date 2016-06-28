package mq

import "log"

//"log"

func TestConsumeGroup() {
	msgChan := make(chan []byte, 1024)
	ConsumeGroup("localhsot:2181/kafak", "testGroup", ConsumeMsgGroup, 0, 0, msgChan)

	for {
		msg := <-msgChan
		log.Println(string(msg))
	}
}
