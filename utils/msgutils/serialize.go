package msgutils

// Serializer serialze and deserialize messages.
type Serializer interface {
	Serialize(Message) ([]byte, error)
	Deserialize([]byte) (Message, error)
}
