package proto

import (
	"encoding/json"
	"xim/utils/msgutils"
)

// JSONSerializer is an implementation of Serializer that handles serializing
// and deserializing JSON encoded payloads.
type JSONSerializer struct {
}

// Serialize marshals the payload into a message.
func (s *JSONSerializer) Serialize(v msgutils.Message) ([]byte, error) {
	return json.Marshal(v)
}

// Deserialize unmarshals the payload into a message.
func (s *JSONSerializer) Deserialize(data []byte) (msg msgutils.Message, err error) {
	msg = new(Msg)
	err = json.Unmarshal(data, msg)
	return
}
