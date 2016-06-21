package mq

import (
	"github.com/Shopify/sarama"
)

var (
	kafkaProducer sarama.AsyncProducer
)

func initKafka(addrs []string) (err error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.NoResponse
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

func handling() {
	successes := kafkaProducer.Successes()
	errors := kafkaProducer.Errors()
	for {
		select {
		case pm := <-successes:
			if pm != nil {
				l.Debug("pub msg success, partition:%d offset:%d key:%v valus:%s", pm.Partition, pm.Offset, pm.Key, pm.Value)
			}
		case err := <-errors:
			if err != nil {
				l.Warning("pub msg error, partition:%d offset:%d key:%v valus:%s error(%v)", err.Msg.Partition, err.Msg.Offset, err.Msg.Key, err.Msg.Value, err.Err)
			}
		}
	}
}
