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

	"math/rand"

	"gopkg.in/webee/turnpike.v2"
)

const (
	maxPending   = 1024
	latencyTopic = "latency"
	procedure    = "procedure"
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
	index     int32   = 1
)

var realm = flag.String("realm", "xchat", "realm")
var debug = flag.Bool("debug", true, "debug mode")
var testing = flag.Bool("testing", true, "testing mode")
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

	//go latencySub(host + ":" + strconv.Itoa(port))
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
	time.Sleep(time.Duration(rand.Intn(*duration)) * time.Microsecond)
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
		// 避免同时发送
		time.Sleep(time.Duration(rand.Intn(*duration)) * time.Second)
		_, err := c.Call(procedure, []interface{}{strconv.Itoa(id)}, nil)
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
	c.Subscribe(topic, OnRecvMsg)
}

func OnRecvMsg(args []interface{}, kwargs map[string]interface{}) {
	details := args[0].(interface{})
	log.Println("recvMsg: ", details)
}

func OnRecvLatencyMsg(args []interface{}, kwargs map[string]interface{}) {
	end = time.Now()
	details := args[0].(string)
	fmt.Println("  endTime:", end, details)

	diff := end.Sub(start).Seconds() * 1000 //ms
	if diff > max {
		max = diff
	}
	if diff < min {
		min = diff
	}
	sum += diff
	ltimes++
	fmt.Printf("..latency: %0.2fms max: %0.2fms min: %0.2fms avg: %0.2fms(%0.2f/%d)\n", diff, max, min, sum/float64(ltimes), sum, ltimes)
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

	err = c.Subscribe(latencyTopic, OnRecvLatencyMsg)
	if err != nil {
		log.Println("latency subscribe failed.", err)
	}

	var i int = 1
	for {
		start = time.Now()
		fmt.Println("startTime:", start, i)
		_, err := c.Call(procedure, []interface{}{latencyTopic + strconv.Itoa(i)}, nil)
		if err != nil {
			log.Println("latency publish failed.", err)
		}
		i++
		time.Sleep(10 * time.Second)
	}
}
