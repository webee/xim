package mq

// topics
const (
	XChatUserStatuses  = "xchat_user_statuses"
	XChatLogsTopic     = "xchat_logs"
	XChatUserMsgsTopic = "xchat_user_msgs"
	XChatCSReqs        = "xchat_cs_reqs"
)

// Publish publish msg to topic.
var (
	Publish             = nilPublish
	PublishBytes        = nilPublishBytes
	PublishBytesWithKey = nilPublishBytesWithKey
)

// InitMQ init message queues.
func InitMQ(kafkaAddrs []string) (close func()) {
	if len(kafkaAddrs) > 0 {
		if err := initKafka(kafkaAddrs); err != nil {
			l.Warning("init kafka failed:", err.Error())
		} else {
			Publish = publishToKafka
			PublishBytes = publishBytesToKafka
			PublishBytesWithKey = publishBytesWithKeyToKafka
			return func() {
				kafkaProducer.Close()
			}
		}
	}
	return func() {}
}

func nilPublish(topic string, msg string) error {
	l.Info("nil publish: %s, %s", topic, msg)
	return nil
}

func nilPublishBytes(topic string, msg []byte) error {
	l.Info("nil publish: %s, %s", topic, string(msg))
	return nil
}

func nilPublishBytesWithKey(topic string, key string, msg []byte) error {
	l.Info("nil publish: %s, %s, %s", topic, key, string(msg))
	return nil
}
