package proto

import (
	"encoding/json"
	"errors"
	"log"
	"xim/utils/msgutils"
)

// Serialization indicates the data serialization format used in a connection.
type Serialization int

// serializations.
const (
	JSONOBJ Serialization = iota
)

// JSONObjSerializer is an implementation of Serializer that handles serializing
// and deserializing JSON Object encoded messages.
type JSONObjSerializer struct {
}

// MsgType hold the msg type.
type MsgType struct {
	Type string `json:"type"`
}

// Serialize marshals the payload into a message.
func (s *JSONObjSerializer) Serialize(m msgutils.Message) ([]byte, error) {
	log.Printf("serialze: %s %+v\n", XIMMsgType(m.MessageType()).String(), m)
	var msg msgutils.Message
	switch x := m.(type) {
	case *Null:
		msg = &struct {
			Type string `json:"type,omitempty"`
			*Null
		}{NULL.String(), x}
	case *Hello:
		msg = &struct {
			Type string `json:"type,omitempty"`
			*Hello
		}{HELLO.String(), x}
	case *Ping:
		msg = &struct {
			Type string `json:"type,omitempty"`
			*Ping
		}{PING.String(), x}
	case *Pong:
		msg = &struct {
			Type string `json:"type,omitempty"`
			*Pong
		}{PONG.String(), x}
	case *Bye:
		msg = &struct {
			Type string `json:"type,omitempty"`
			*Bye
		}{BYE.String(), x}
	case *Put:
		var uid interface{}
		if x.UID != 0 {
			uid = x.UID
		}
		msg = &struct {
			Type string `json:"type,omitempty"`
			*Put
			UID interface{} `json:"uid,omitempty"`
		}{PUT.String(), x, uid}
	case *Push:
		var uid interface{}
		if x.UID != 0 {
			uid = x.UID
		}
		msg = &struct {
			Type string `json:"type,omitempty"`
			*Push
			UID interface{} `json:"uid,omitempty"`
		}{PUSH.String(), x, uid}
	case *Reply:
		var uid interface{}
		if x.UID != 0 {
			uid = x.UID
		}
		msg = &struct {
			Type string `json:"type,omitempty"`
			*Reply
			UID interface{} `json:"uid,omitempty"`
		}{REPLY.String(), x, uid}
	case *Register:
		msg = &struct {
			Type string `json:"type,omitempty"`
			*Register
		}{REGISTER.String(), x}
	case *Unregister:
		msg = &struct {
			Type string `json:"type,omitempty"`
			*Unregister
		}{UNREGISTER.String(), x}
	default:
		return nil, errors.New("unkown message type")
	}
	return json.Marshal(msg)
}

// Deserialize unmarshals the payload into a message.
func (s *JSONObjSerializer) Deserialize(data []byte) (msg msgutils.Message, err error) {
	log.Println("deserialize: ", string(data))
	msgType := MsgType{}
	err = json.Unmarshal(data, &msgType)
	if err != nil {
		return nil, err
	}

	switch msgType.Type {
	case NULL.String():
		return NULL.New(), nil
	case HELLO.String():
		msg := HELLO.New()
		return msg, json.Unmarshal(data, msg)
	case PING.String():
		return PING.New(), nil
	case PONG.String():
		return PONG.New(), nil
	case BYE.String():
		return BYE.New(), nil
	case PUT.String():
		msg := PUT.New()
		return msg, json.Unmarshal(data, msg)
	case PUSH.String():
		msg := PUSH.New()
		return msg, json.Unmarshal(data, msg)
	case REPLY.String():
		msg := REPLY.New()
		return msg, json.Unmarshal(data, msg)
	case REGISTER.String():
		msg := REGISTER.New()
		return msg, json.Unmarshal(data, msg)
	case UNREGISTER.String():
		msg := UNREGISTER.New()
		return msg, json.Unmarshal(data, msg)
	default:
		return nil, errors.New("unkown message type")
	}
}
