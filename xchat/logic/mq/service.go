package mq

// topics
const (
	XChatLogsTopic     = "xchat_logs"
	XChatUserMsgsTopic = "xchat_user_msgs"
)

// InitMQ init message queues.
func InitMQ(kafkaAddrs []string) (close func()) {
	initKafka(kafkaAddrs)
	return func() {
		kafkaProducer.Close()
	}
}

// Publish publish msg to topic.
func Publish(topic string, msg string) error {
	return publishToKafka(topic, msg)
}
