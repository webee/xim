package msgutils

import "encoding/json"

// Serializer serialze and deserialize messages.
type Serializer interface {
	Serialize(Message) ([]byte, error)
	Deserialize([]byte) (Message, error)
}

// JSONSerializer is an implementation of Serializer that handles serializing
// and deserializing JSON encoded payloads.
type JSONSerializer struct {
}

// Serialize marshals the payload into a message.
func (s *JSONSerializer) Serialize(m Message) ([]byte, error) {
	return json.Marshal(m)
}

// Deserialize unmarshals the payload into a message.
func (s *JSONSerializer) Deserialize(data []byte) (msg Message, err error) {
	msg = new(map[string]interface{})
	err = json.Unmarshal(data, msg)
	return
}
