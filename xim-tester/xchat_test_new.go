package main

import (
	"flag"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"github.com/dgrijalva/jwt-go"
	"gopkg.in/jcelliott/turnpike.v2"
	"math/rand"
	"xim/xchat/broker/mid"
)

const (
	mySigningKey = "demo app user key."
)

var (
	connected int64
	pending   int64
	failed    int64
	run       bool = true
)

var realm = flag.String("realm", "xchat", "realm")
var addr = flag.String("addr", "localhost:3699", "wamp server addr")
var times = flag.Int("times", 10, "send msg times")
var concurrent = flag.Int64("concurrent", 1, "concurrent users")
var timeout = flag.Duration("timeout", 30*time.Second, "timeout for recv")
var rate = flag.Int("rate", 1, "the rate between user and channel")
var duration = flag.Int("duration", 30, "duration between msgs")
var maxpending = flag.Int64("pending", 10, "max concurrent connecting")

type MyClient turnpike.Client

func main() {
	flag.Parse()
	rand.Seed(int64(time.Now().Second()))
	go func() {
		start := time.Now()
		for {
			fmt.Printf("client elapsed=%0.0fs pending=%d connected=%d failed=%d\n", time.Now().Sub(start).Seconds(),
				atomic.LoadInt64(&pending), atomic.LoadInt64(&connected), atomic.LoadInt64(&failed))
			time.Sleep(1 * time.Second)
		}
	}()
	exit := make(chan bool, *concurrent)

	var i int = 1
	for {
		if atomic.LoadInt64(&connected)+atomic.LoadInt64(&pending) < *concurrent &&
			atomic.LoadInt64(&pending) < *maxpending && run {
			atomic.AddInt64(&pending, 1)
			go newClient(i/(*rate), exit, addr)
			i++
		} else {
			time.Sleep(100 * time.Millisecond)
		}
	}

	for i := int64(0); i < *concurrent; i++ {
		<-exit
	}
}

func createToken(user string, valid time.Duration) (string, error) {
	// Create the token
	token := jwt.New(jwt.SigningMethodHS256)
	// Set some claims
	token.Claims["user"] = "test"
	token.Claims["username"] = "test"
	token.Claims["exp"] = time.Now().Add(valid).Unix()
	// Sign and get the complete encoded token as a string
	tokenString, err := token.SignedString([]byte(mySigningKey))

	return tokenString, err
}

func AuthFunc(hello map[string]interface{}, c map[string]interface{}) (string, map[string]interface{}, error) {
	log.Println(hello, c)
	challenge, ok := c["challenge"].(string)
	if ok {
		log.Fatal("no challenge data recevied", challenge)
	}
	token, err := createToken("test", 72*time.Hour)
	if err != nil {
		return "", nil, err
	}
	return token, nil, nil
}

func newClient(id int, exit chan bool, addr *string) {
	c, err := turnpike.NewWebsocketClient(turnpike.JSON, "ws://"+*addr+"/ws", nil)
	if err != nil {
		log.Println(id, "new websocket client failed.", err)
		atomic.AddInt64(&pending, -1)
		atomic.AddInt64(&failed, 1)
		return
	}
	atomic.AddInt64(&pending, -1)
	atomic.AddInt64(&connected, 1)

	c.ReceiveTimeout = *timeout
	c.Auth = map[string]turnpike.AuthFunc{"jwt": AuthFunc}

	ret, err := c.JoinRealm(*realm, nil)
	if err != nil {
		log.Println(id, "join realm failed.", err)
		atomic.AddInt64(&connected, -1)
		atomic.AddInt64(&failed, 1)
		return
	} else {
		log.Println("joinRealm", ret)
	}

	var channelID turnpike.ID = 1
	var user string = "test"

	// ping and get session id
	session, err := c.Call(mid.URIXChatPing, nil, map[string]interface{}{
		"detail": map[string]interface{}{
			"session": id,
			"user":    user,
		},
	})
	if err != nil {
		log.Println("ping failed.", err)
	} else {
		log.Println("ping return", session)
	}
	tmpId, ok := session.Arguments[1].(float64)
	if !ok {
		log.Println("get sesssionId failed.", err)
	}
	sessionId := uint64(tmpId)

	topic := fmt.Sprintf(mid.URIXChatUserMsg, sessionId)
	go recvMsg(topic, c)

	for i := 0; i < *times && run; i++ {
		// 避免同时发送
		time.Sleep(time.Duration(rand.Intn(*duration)) * time.Second)
		result, err := c.Call(mid.URIXChatSendMsg, []interface{}{channelID, "hello, xchat"}, map[string]interface{}{
			"detail": map[string]interface{}{
				"session": channelID,
				"user":    user,
			},
		})
		log.Println("rpc called. ret:", result)
		if err != nil {
			log.Println("Error Sending message", err)
			break
		}
	}
	atomic.AddInt64(&connected, -1)
	atomic.AddInt64(&failed, 1)
	log.Println(id, "client exit")
	c.Close()
	exit <- true
}

func recvMsg(topic string, c *turnpike.Client) {
	log.Printf("sub topic{%s}\n", topic)
	err := c.Subscribe(topic, OnRecvMsg)
	if err != nil {
		log.Println("subscribe failed.", err)
	} else {
		log.Println("......................................subscribe success.")
	}
}

func OnRecvMsg(args []interface{}, kwargs map[string]interface{}) {
	details := args[0].(interface{})
	log.Println("recvMsg: ", details)
}
