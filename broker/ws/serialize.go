package ws

import (
	"encoding/json"
	"xim/broker/proto"
	"xim/utils/msgutils"
)

// ProtoJSONSerializer is an implementation of Serializer that handles serializing
// and deserializing JSON encoded payloads.
type ProtoJSONSerializer struct {
}

// Serialize marshals the payload into a message.
func (s *ProtoJSONSerializer) Serialize(v msgutils.Message) ([]byte, error) {
	return json.Marshal(v)
}

// Deserialize unmarshals the payload into a message.
func (s *ProtoJSONSerializer) Deserialize(data []byte) (msg msgutils.Message, err error) {
	msg = new(proto.Msg)
	err = json.Unmarshal(data, msg)
	return
}
