package mq

// topics
const (
	XChatLogsTopic     = "xchat_logs"
	XChatUserMsgsTopic = "xchat_user_msgs"
	XChatCSReqs        = "xchat_cs_reqs"
)

// InitMQ init message queues.
func InitMQ(kafkaAddrs []string) (close func()) {
	if err := initKafka(kafkaAddrs); err != nil {
		l.Warning("init kafka failed:", err.Error())
	}

	return func() {
		kafkaProducer.Close()
	}
}

// Publish publish msg to topic.
func Publish(topic string, msg string) error {
	return publishToKafka(topic, msg)
}
