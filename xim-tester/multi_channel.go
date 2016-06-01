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

	"github.com/gorilla/websocket"
)

var (
	addr     = flag.String("addr", "localhost:2880", "http service address")
	webaddr  = flag.String("webaddr", "localhost:6980", "web server addr")
	nums     = flag.Int("nums", 1, "online users on the same time.")
	username = flag.String("username", "test", "username")
	password = flag.String("password", "test1234", "password")
	msg      = flag.String("msg", "{}", "msg to send.")
	interval = flag.Duration("interval", 100*time.Microsecond, "msg send interval")
)

const (
	rawHeaderLen = uint16(16)
	heart        = 10 * time.Second
)

var (
	sendTimes   int64
	failedTimes int64
	recvTimes   int64
	token       string
	userToken   string
	chanel      string
)

// Token is token.
type Token struct {
	Token string `json:"token"`
}

type ChannelStruct struct {
	App     string `json:"app"`
	Channel string `json:"channel"`
	Created string `json:"created"`
	Tag     string `json:"tag"`
	Updated string `json:"updated"`
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

func getNewChannel(addr, token, user string) string {
	reader := strings.NewReader(fmt.Sprintf(`{"tag": "test", "pubs": ["%s"], "subs": ["%s"]}`, user, user))
	response, err := http.DefaultClient.Post("http://"+addr+"/xim/app/channels/?jwt="+token,
		"application/json; charset=utf-8", reader)
	if err != nil {
		log.Fatal("post failed!")
	}
	content := make([]byte, 1024)
	n, err := response.Body.Read(content)
	if err != nil {
		log.Fatal("read body failed.")
	}
	response.Body.Close()
	var t ChannelStruct
	json.Unmarshal(content[:n], &t)
	return t.Channel
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()

	// get token from web
	token = getNewToken(*webaddr, *username, *password)
	userToken = getNewUserToken(token, *webaddr, *username, "2592000")
	chanel = getNewChannel(*webaddr, token, *username)

	for i := 0; i < *nums; i++ {
		go startClient(i)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}

func startClient(id int) {
	u := url.URL{Scheme: "ws", Host: *addr, Path: "/ws"}

	headers := http.Header{}
	headers.Add("Authorization", "Bearer "+userToken)
	c, _, err := websocket.DefaultDialer.Dial(u.String(), headers)
	if err != nil {
		log.Printf("#%d\terror: %s\n", id, err.Error())
		return
	}
	log.Printf("#%d\t connected.", id)

	quit := make(chan bool, 1)
	defer close(quit)

	msgToSend := []byte(*msg)
	heartbeat := time.After(30 * time.Second)
	// writer
	go func() {
		for {
			select {
			case <-heartbeat:
				err := c.WriteMessage(websocket.TextMessage, []byte("{}"))
				if err != nil {
					log.Printf("#%d\terror: %s\n", id, err.Error())
					return
				}
				heartbeat = time.After(30 * time.Second)
			case <-time.After(*interval):
				err := c.WriteMessage(websocket.TextMessage, msgToSend)
				if err != nil {
					log.Printf("#%d\terror: %s\n", id, err.Error())
					return
				}
			}
			select {
			case <-quit:
				return
			default:
			}
		}
	}()

	// reader
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Printf("#%d\terror: %s\n", id, err.Error())
			return
		}
		log.Printf("#%d\t: %s\n", id, message)
	}
}
