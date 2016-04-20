package main

import (
	"fmt"
	"time"
	"xim/dispatcher/msgchan"
)

func main() {
	msgChan := msgchan.NewMsgChan()
	for i := 0; i < 10000; i++ {
		go msgChan.Put(fmt.Sprintf("#%d", i))
	}

	time.Sleep(100 * time.Millisecond)
	msgChan.Close()
	fmt.Println("total:", msgChan.Count())
}
