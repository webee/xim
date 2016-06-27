package mq

import (
	"github.com/Shopify/sarama"
	"github.com/wvanbergen/kafka/consumergroup"
	"github.com/wvanbergen/kazoo-go"
)

// kafka topic & group config
const (
	XchatLogTopic = "xchat_logs"
	XchatMsgTopic = "xchat_user_msgs"

	ConsumeMsgGroup = "consume_msg_group"
	ConsumeLogGroup = "consume_log_group"
)

// ConsumeGroup consume kafka message in group
func ConsumeGroup(zkaddr string, group, topic string, index, offset int, msgChan chan []byte) error {
	config := consumergroup.NewConfig()
	config.Offsets.Initial = sarama.OffsetOldest
	//config.Offsets.CommitInterval = 30 * time.Second

	l.Info("zkAddr: %s", zkaddr)
	var zkNodes []string
	zkNodes, config.Zookeeper.Chroot = kazoo.ParseConnectionString(zkaddr)

	cg, err := consumergroup.JoinConsumerGroup(group, []string{topic}, zkNodes, config)
	if err != nil {
		l.Warning("JoinConsumerGroup failed. %s", err.Error())
		return err
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
