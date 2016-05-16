package ws

import "xim/broker/proto"

import "encoding/json"

// Serializer serialze and deserialize proto messages.
type Serializer interface {
	Serialize(interface{}) ([]byte, error)
	Deserialize([]byte) (*proto.Msg, error)
}

// JSONSerializer is an implementation of Serializer that handles serializing
// and deserializing JSON encoded payloads.
type JSONSerializer struct {
}

// Serialize marshals the payload into a message.
func (s *JSONSerializer) Serialize(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// Deserialize unmarshals the payload into a message.
func (s *JSONSerializer) Deserialize(data []byte) (msg *proto.Msg, err error) {
	msg = new(proto.Msg)
	err = json.Unmarshal(data, msg)
	return
}
