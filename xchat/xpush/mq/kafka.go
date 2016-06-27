package mq

import (
	"github.com/Shopify/sarama"
	"github.com/wvanbergen/kafka/consumergroup"
	"github.com/wvanbergen/kazoo-go"
)

const (
	XCHAT_LOG_TOPIC = "xchat_logs"
	XCHAT_MSG_TOPIC = "xchat_user_msgs"

	CONSUME_MSG_GROUP = "consume_msg_group"
	CONSUME_LOG_GROUP = "consume_log_group"
)

// 同步produce kafka消息
func SyncProduce(addr []string, topic, msg string, partition int32) error {
	config := sarama.NewConfig()
	config.Producer.Partitioner = sarama.NewManualPartitioner
	producer, err := sarama.NewSyncProducer(addr, config)
	if err != nil {
		l.Error("NewSyncProducer failed. %v", err)
		return err
	}
	defer func() {
		if err := producer.Close(); err != nil {
			l.Error("producer close failed. %v", err)
		}
	}()

	m := &sarama.ProducerMessage{Topic: topic, Partition: partition,
		Value: sarama.StringEncoder(msg)}
	partition, offset, err := producer.SendMessage(m)
	if err != nil {
		l.Error("produce msg failed. %v", err)
		return err
	} else {
		l.Info("msg send to partion %d offset %d %s.", partition, offset, msg)
	}
	return nil
}

// 异步产生kafka消息
func AsyncProduce(addr []string, topic string, msgChan chan []byte) error {
	config := sarama.NewConfig()
	producer, err := sarama.NewAsyncProducer(addr, config)
	if err != nil {
		l.Error("NewAsyncProducer failed. %v", err)
		return err
	}
	defer func() {
		if err := producer.Close(); err != nil {
			l.Error("producer close failed., %v", err)
		}
	}()

	for {
		select {
		case msg := <-msgChan:
			producer.Input() <- &sarama.ProducerMessage{Topic: topic, Value: sarama.StringEncoder(msg)}
		case err := <-producer.Errors():
			l.Error("failed to produce message. %v", err)
		}
	}
	return nil
}

// 消费kafka消息
func Consume(addr []string, topic string, partition int32, offset int64, chanMsg chan []byte) error {
	config := sarama.NewConfig()
	//config.Consumer.Offsets.Initial = sarama.OffsetOldest
	consumer, err := sarama.NewConsumer(addr, config)
	if err != nil {
		l.Error("NewConsumer failed. %v", err)
		return err
	}
	//
	//id, err := consumer.Partitions(topic)
	//if err != nil {
	//	log.Println("consumer partitions failed.", err)
	//} else {
	//	log.Println("partitions", id)
	//}

	partitionConsumer, err := consumer.ConsumePartition(topic, partition, offset)
	if err != nil {
		l.Error("ConsumePartition failed. %v", err)
		return err
	}

	defer func() {
		if consumer.Close(); err != nil {
			l.Error("close consume failed. %v", err)
		}
	}()

	for {
		select {
		case msg := <-partitionConsumer.Messages():
			l.Debug("key: %s, value: %s, topic: %s, partition: %d, offset: %d",
				string(msg.Key), string(msg.Value), msg.Topic, msg.Partition, msg.Offset)
			chanMsg <- msg.Value
		}
	}

	return nil
}

// 按组消费kafka消息
func ConsumeGroup(zkaddr string, group, topic string, index, offset int, msgChan chan []byte) error {
	config := consumergroup.NewConfig()
	config.Offsets.Initial = sarama.OffsetOldest
	//config.Offsets.CommitInterval = 30 * time.Second

	l.Info("zkAddr", zkaddr)
	var zkNodes []string
	zkNodes, config.Zookeeper.Chroot = kazoo.ParseConnectionString(zkaddr)

	cg, err := consumergroup.JoinConsumerGroup(group, []string{topic}, zkNodes, config)
	if err != nil {
		l.Error("JoinConsumerGroup failed. %v", err)
		return err
	}

	defer func() {
		if err := cg.Close(); err != nil {
			l.Error("cg close failed. %v", err)
		}
	}()

	for {
		select {
		case msg := <-cg.Messages():
			msgChan <- msg.Value
			l.Debug("{index:%d} key: %s, value: %s, topic: %s, partition: %d, offset: %d", index,
				string(msg.Key), string(msg.Value), msg.Topic, msg.Partition, msg.Offset)
			cg.CommitUpto(msg)
		}
	}

	return nil
}
