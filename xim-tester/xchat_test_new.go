package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"sync/atomic"
	"time"

	"xim/xchat/broker/mid"

	"github.com/dgrijalva/jwt-go"
	"gopkg.in/jcelliott/turnpike.v2"
)

const (
	mySigningKey = "demo app user key."
)

var (
	connected int64
	pending   int64
	failed    int64
	run       = true
)

var realm = flag.String("realm", "xchat", "realm")
var addr = flag.String("addr", "localhost:3699", "wamp server addr")
var times = flag.Int("times", 10, "send msg times")
var concurrent = flag.Int64("concurrent", 1, "concurrent users")
var timeout = flag.Duration("timeout", 30*time.Second, "timeout for recv")
var rate = flag.Int("rate", 1, "the rate between user and channel")
var duration = flag.Int("duration", 30, "duration between msgs")
var maxpending = flag.Int64("pending", 10, "max concurrent connecting")

// MyClient is a wamp client.
type MyClient turnpike.Client

func init() {
	rand.Seed(time.Now().Unix())
}

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

	i := int(1)
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
}

func createToken(user string, valid time.Duration) (string, error) {
	// Create the token
	token := jwt.New(jwt.SigningMethodHS256)
	// Set some claims
	token.Claims["user"] = user
	token.Claims["username"] = user
	token.Claims["exp"] = time.Now().Add(valid).Unix()
	// Sign and get the complete encoded token as a string
	tokenString, err := token.SignedString([]byte(mySigningKey))

	return tokenString, err
}

func genAuthFunc(user string) turnpike.AuthFunc {
	return func(hello map[string]interface{}, c map[string]interface{}) (string, map[string]interface{}, error) {
		log.Println(hello, c)
		challenge, ok := c["challenge"].(string)
		if ok {
			log.Fatal("no challenge data recevied", challenge)
		}
		token, err := createToken(user, 72*time.Hour)
		if err != nil {
			return "", nil, err
		}
		return token, nil, nil
	}
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

	defer func() {
		atomic.AddInt64(&connected, -1)
		atomic.AddInt64(&failed, 1)
	}()

	c.ReceiveTimeout = *timeout
	c.Auth = map[string]turnpike.AuthFunc{"jwt": genAuthFunc(strconv.Itoa(id))}

	ret, err := c.JoinRealm(*realm, nil)
	if err != nil {
		log.Println(id, "join realm failed.", err)
		return
	}
	log.Println("joinRealm", ret)

	user := "test"

	// ping and get session id
	session, err := c.Call(mid.URIXChatPing, nil, map[string]interface{}{
		"detail": map[string]interface{}{
			"session": id,
			"user":    user,
		},
	})
	if err != nil {
		log.Println("ping failed.", err)
		return
	}
	log.Println("ping return", session)
	tmpID, ok := session.Arguments[1].(float64)
	if !ok {
		log.Println("get sesssionId failed.", err)
	}
	sessionID := uint64(tmpID)

	topic := fmt.Sprintf(mid.URIXChatUserMsg, sessionID)
	recvMsg(topic, c)

	chatID := (id-1)/60 + 1
	for i := 0; i < *times && run; i++ {
		// 避免同时发送
		time.Sleep(time.Duration(rand.Intn(*duration)) * time.Second)
		result, err := c.Call(mid.URIXChatSendMsg, []interface{}{chatID, "hello, xchat"}, map[string]interface{}{})
		log.Println("rpc called. ret:", result)
		if err != nil {
			log.Println("Error Sending message", err)
			break
		}
	}
	log.Println(id, "client exit")
	c.Close()
	exit <- true
}

func recvMsg(topic string, c *turnpike.Client) {
	log.Printf("sub topic{%s}\n", topic)
	err := c.Subscribe(topic, nnRecvMsg)
	if err != nil {
		log.Println("subscribe failed.", err)
	} else {
		log.Println("......................................subscribe success.")
	}
}

func nnRecvMsg(args []interface{}, kwargs map[string]interface{}) {
	details := args[0].(interface{})
	log.Println("recvMsg: ", details)
}
