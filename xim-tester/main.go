package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"time"

	"xim/apps/xchat/mid"
	"xim/broker/proto"
	"xim/utils/msgutils"
)

var (
	addr     = flag.String("addr", "localhost:2880", "http service address")
	webaddr  = flag.String("webaddr", "localhost:6980", "web server addr")
	nums     = flag.Int("nums", 1, "online users on the same time.")
	username = flag.String("username", "test", "username")
	password = flag.String("password", "test1234", "password")
	interval = flag.Duration("interval", 100*time.Millisecond, "msg send interval")
	channel  = flag.String("channel", "NB845YNO", "channel to send")
	msg      = flag.String("msg", "hello.", "msg to send.")
)

var (
	token     string
	userToken string
)

// Token is token.
type Token struct {
	Token string `json:"token"`
}

func getNewToken(addr, user, password string) string {
	reader := strings.NewReader(fmt.Sprintf(`{"username": "%s", "password": "%s"}`, user, password))
	response, err := http.DefaultClient.Post("http://"+addr+"/xim/app.new_token",
		"application/json; charset=utf-8", reader)
	if err != nil {
		log.Fatal("post failed")
	}
	content := make([]byte, 1024)
	n, err := response.Body.Read(content)
	if err != nil {
		log.Fatal("read body failed.")
	}
	response.Body.Close()
	var t Token
	json.Unmarshal(content[:n], &t)
	return t.Token
}

func getNewUserToken(token, addr, user, expire string) string {
	reader := strings.NewReader(fmt.Sprintf(`{"user": "%s", "expire": "%s"}`, user, expire))
	response, err := http.DefaultClient.Post("http://"+addr+"/xim/app.new_user_token?jwt="+token,
		"application/json; charset=utf-8", reader)
	if err != nil {
		log.Fatal("post failed.")
	}
	content := make([]byte, 1024)
	n, err := response.Body.Read(content)
	if err != nil {
		log.Fatal("read body failed.")
	}
	response.Body.Close()
	var t Token
	json.Unmarshal(content[:n], &t)
	return t.Token
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()

	// get token from web
	token = getNewToken(*webaddr, *username, *password)
	userToken = getNewUserToken(token, *webaddr, *username, "2592000")

	for i := 0; i < *nums; i++ {
		go startClient(i)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}

func startClient(id int) {
	u := url.URL{Scheme: "ws", Host: *addr, Path: "/ws"}

	t, err := mid.GetWSTranseiver(u.String(), userToken, 10)
	if err != nil {
		log.Printf("#%d\terror: %s\n", id, err.Error())
		return
	}
	log.Printf("#%d\t connected.", id)

	msgController := msgutils.NewMsgController(t, genMsgHandler(id), nil)
	defer msgController.Close()
	msgController.Start()

	// writer
	for {
		select {
		case <-time.After(*interval):
			reply, err := msgController.SyncSend(&proto.Put{
				Channel: fmt.Sprintf("c%d", id),
				Msg:     *msg,
			})
			if err != nil {
				log.Printf("#%d\terror: %s\n", id, err.Error())
				return
			}
			log.Printf("#%d reply\t: %+v\n", id, reply)
		}
	}
}

func genMsgHandler(id int) msgutils.MessageHandler {
	return func(msg msgutils.Message) {
		log.Printf("#%d\t: %+v\n", id, msg)
	}
}
