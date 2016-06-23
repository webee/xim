package mq

// topics
const (
	XChatLogsTopic     = "xchat_logs"
	XChatUserMsgsTopic = "xchat_user_msgs"
	XChatCSReqs        = "xchat_cs_reqs"
)

// Publish publish msg to topic.
var (
	Publish = nilPublish
)

// InitMQ init message queues.
func InitMQ(kafkaAddrs []string) (close func()) {
	if err := initKafka(kafkaAddrs); err != nil {
		l.Warning("init kafka failed:", err.Error())
	} else {
		Publish = publishToKafka
	}

	return func() {
		kafkaProducer.Close()
	}
}

func nilPublish(topic string, msg string) error {
	l.Info("nil publish: %s, %s", topic, msg)
	return nil
}
