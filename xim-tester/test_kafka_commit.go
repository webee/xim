package main

import (
	"time"
	"xim/xchat/logic/logger"
	"xim/xchat/xpush/mq"

	"flag"

	"github.com/Shopify/sarama"
	"github.com/wvanbergen/kafka/consumergroup"
	"github.com/wvanbergen/kazoo-go"
)

var (
	l              = logger.Logger
	commitInterval = flag.Duration("commitInterval", 10*time.Second, "kafka group commit iterval.")
	topic          = flag.String("topic", mq.XchatMsgTopic, "the kafka topic")
	group          = flag.String("group", mq.ConsumeMsgGroup, "the kafka consume group")
)

func main() {
	flag.Parse()
	zkaddr := "localhost:2181/kafka"
	group := *group
	topic := *topic
	index := 0
	msgChan := make(chan []byte, 1024)
	config := consumergroup.NewConfig()
	config.Offsets.Initial = sarama.OffsetOldest
	config.Offsets.CommitInterval = *commitInterval

	l.Info("zkAddr: %s", zkaddr)
	var zkNodes []string
	zkNodes, config.Zookeeper.Chroot = kazoo.ParseConnectionString(zkaddr)

	cg, err := consumergroup.JoinConsumerGroup(group, []string{topic}, zkNodes, config)
	if err != nil {
		l.Warning("JoinConsumerGroup failed. %s", err.Error())
		return
	}

	defer func() {
		if err := cg.Close(); err != nil {
			l.Warning("cg close failed. %s", err.Error())
		}
	}()

	for {
		select {
		case msg := <-cg.Messages():
			msgChan <- msg.Value
			l.Debug("{index:%d} key: %s, value: %s, topic: %s, partition: %d, offset: %d", index,
				string(msg.Key), string(msg.Value), msg.Topic, msg.Partition, msg.Offset)
			err := cg.CommitUpto(msg)
			if err != nil {
				l.Warning("consumeGroup.CommitUpto failed. %s", err.Error())
			}
		}
	}
}
