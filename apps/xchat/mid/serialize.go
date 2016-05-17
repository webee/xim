package mid

import "encoding/json"

// Serializer serialze and deserialize proto messages.
type Serializer interface {
	Serialize(interface{}) ([]byte, error)
	Deserialize([]byte) (map[string]interface{}, error)
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
func (s *JSONSerializer) Deserialize(data []byte) (msg map[string]interface{}, err error) {
	msg = make(map[string]interface{})
	err = json.Unmarshal(data, &msg)
	return
}
