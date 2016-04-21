package main

import (
	"fmt"
	"time"
	"xim/dispatcher/msgchan"
)

func handler(n interface{}) interface{} {
	return n.(int) + 1
}

func print(s interface{}) error {
	_, err := fmt.Println(s.(int))
	return err
}

func main() {
	msgChan2 := msgchan.NewMsgChannel("YYY", 10, handler, msgchan.NewMsgChannelHandlerDownStream("printer", print), 10*time.Minute)
	msgChan := msgchan.NewMsgChannel("XXX", 10, handler, msgChan2, 10*time.Minute)
	for i := 0; i < 10; i++ {
		msgChan.Put(i)
	}

	time.Sleep(100 * time.Millisecond)
	msgChan.Close()
}
