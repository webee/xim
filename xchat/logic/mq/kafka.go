package mq

import (
	"github.com/Shopify/sarama"
)

var (
	kafkaProducer sarama.AsyncProducer
)

func initKafka(addrs []string) (err error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForLocal
	config.Producer.Partitioner = sarama.NewHashPartitioner
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	kafkaProducer, err = sarama.NewAsyncProducer(addrs, config)
	if err != nil {
		return
	}

	go handling()
	return
}

func publishToKafka(topic string, msg string) error {
	kafkaProducer.Input() <- &sarama.ProducerMessage{Topic: topic, Value: sarama.StringEncoder(msg)}
	return nil
}

func publishBytesToKafka(topic string, msg []byte) error {
	kafkaProducer.Input() <- &sarama.ProducerMessage{Topic: topic, Value: sarama.ByteEncoder(msg)}
	return nil
}

func publishBytesWithKeyToKafka(topic string, key string, msg []byte) error {
	kafkaProducer.Input() <- &sarama.ProducerMessage{Topic: topic, Key: sarama.StringEncoder(key), Value: sarama.ByteEncoder(msg)}
	return nil
}

func handling() {
	successes := kafkaProducer.Successes()
	errors := kafkaProducer.Errors()
	for {
		select {
		case pm := <-successes:
			if pm != nil {
				l.Debug("pub msg success, topic:%s, partition:%d offset:%d key:%v values:%s", pm.Topic, pm.Partition, pm.Offset, pm.Key, pm.Value)
			}
		case err := <-errors:
			if err != nil {
				l.Warning("pub msg error, topic:%s, partition:%d offset:%d key:%v values:%s error(%v)", err.Msg.Topic, err.Msg.Partition, err.Msg.Offset, err.Msg.Key, err.Msg.Value, err.Err)
			}
		}
	}
}
