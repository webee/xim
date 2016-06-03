package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync/atomic"
	"syscall"
	"time"

	"gopkg.in/jcelliott/turnpike.v2"
	"math/rand"
)

const (
	maxPending   = 1024
	latencyTopic = "latency"
)

var (
	connected int64
	pending   int64
	failed    int64
	run       bool = true
	start     time.Time
	end       time.Time
	max       float64 = 0.0
	min       float64 = 1000000.0
	sum       float64 = 0
	ltimes    int32   = 0
)
var userkey = flag.String("userkey", "userkey", "user key")
var realm = flag.String("realm", "xchat", "realm")
var debug = flag.Bool("debug", true, "debug mode")
var testing = flag.Bool("testing", true, "testing mode")
var endpoint = flag.String("endpoint", "/ws", "wamp router websocket url endpoint.")
var addr = flag.String("addr", "localhost:3699", "wamp server addr")
var times = flag.Int("times", 10, "send msg times")
var concurrent = flag.Int64("concurrent", 1, "concurrent users")
var timeout = flag.Duration("timeout", 30*time.Second, "timeout for recv")
var duration = flag.Int("duration", 120, "random duration seconds between msgs")
var rate = flag.Int("rate", 1, "the rate between user and channel")

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

	host, _, err := net.SplitHostPort(*addr)
	if err != nil {
		log.Fatalln("split addr failed", err)
	}

	port := 20000
	i := 0

	go latencySub(host + ":" + strconv.Itoa(port))
	go latencyPub(host + ":" + strconv.Itoa(port))
	for {
		if atomic.LoadInt64(&connected)+atomic.LoadInt64(&pending) < *concurrent &&
			atomic.LoadInt64(&pending) < maxPending && run {
			if i > 0 && i%50000 == 0 {
				port++
			}
			atomic.AddInt64(&pending, 1)
			go newClient(i/(*rate), exit, host+":"+strconv.Itoa(port))
			i++
		} else {
			time.Sleep(100 * time.Millisecond)
		}
	}

	setupSignal()

	for i := int64(0); i < *concurrent; i++ {
		<-exit
	}
}

func newClient(id int, exit chan bool, addr string) {
	// 避免同时发起连接
	time.Sleep(time.Duration(rand.Intn(*duration)) * time.Second)
	c, err := turnpike.NewWebsocketClient(turnpike.JSON, "ws://"+addr+"/ws", nil)
	if err != nil {
		log.Println(id, "new websocket client failed.", err)
		atomic.AddInt64(&pending, -1)
		atomic.AddInt64(&failed, 1)
		return
	}
	atomic.AddInt64(&pending, -1)
	atomic.AddInt64(&connected, 1)

	c.ReceiveTimeout = *timeout

	_, err = c.JoinRealm(*realm, nil)
	if err != nil {
		log.Println(id, "join realm failed.", err)
		atomic.AddInt64(&connected, -1)
		atomic.AddInt64(&failed, 1)
		return
	}
	topic := "topic:" + strconv.Itoa(id)
	recvMsg(topic, c)
	for i := 0; i < *times && run; i++ {
		err = c.Publish(topic, []interface{}{id}, nil)
		if err != nil {
			log.Println("Error Sending message", err)
			break
		}
		// 避免同时发送
		time.Sleep(time.Duration(rand.Intn(*duration)) * time.Second)
	}
	atomic.AddInt64(&connected, -1)
	atomic.AddInt64(&failed, 1)
	log.Println(id, "client exit")
	c.Close()
	exit <- true
}

func recvMsg(topic string, c *turnpike.Client) {
	c.Subscribe(topic, OnRecvMsg)
}

func OnRecvMsg(args []interface{}, kwargs map[string]interface{}) {
	//details := args[0].(interface{})
	//	log.Println("recvMsg: ", details)
}

func OnRecvLatencyMsg(args []interface{}, kwargs map[string]interface{}) {
	end = time.Now()
	diff := end.Sub(start).Seconds() * 1000 //ms
	if diff > max {
		max = diff
	}
	if diff < min {
		min = diff
	}
	sum += diff
	ltimes++
	fmt.Printf("..latency: %0.2fms max: %0.2fms min: %0.2fms avg: %0.2fms(%0.2f/%d)\n",
		diff, max, min, sum/float64(ltimes), sum, ltimes)
	details := args[0].(interface{})
	log.Println("recvMsg: ", details)
}

// setupSignal register signals handler and waiting for.
func setupSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	for {
		s := <-c
		log.Println("get a signal: ", s.String())
		switch s {
		case os.Interrupt, syscall.SIGTERM:
			run = false
			return
		default:
			return
		}
	}
}

func latencyPub(addr string) {
	c, err := turnpike.NewWebsocketClient(turnpike.JSON, "ws://"+addr+"/ws", nil)
	if err != nil {
		log.Println("latency websocket client failed.", err)
		return
	}
	c.ReceiveTimeout = *timeout

	_, err = c.JoinRealm(*realm, nil)
	if err != nil {
		log.Println("latency join realm failed.", err)
		return
	}

	for {
		err := c.Publish(latencyTopic, []interface{}{1}, nil)
		if err != nil {
			log.Println("latency publish failed.", err)
		}
		start = time.Now()
		time.Sleep(10 * time.Second)
	}
}

func latencySub(addr string) {
	c, err := turnpike.NewWebsocketClient(turnpike.JSON, "ws://"+addr+"/ws", nil)
	if err != nil {
		log.Println("latency websocket client failed.", err)
		return
	}
	c.ReceiveTimeout = *timeout

	_, err = c.JoinRealm(*realm, nil)
	if err != nil {
		log.Println("latency join realm failed.", err)
		return
	}

	err = c.Subscribe(latencyTopic, OnRecvLatencyMsg)
	if err != nil {
		log.Println("latency subscribe failed.", err)
	}
}
