package rpcservice

import (
	"encoding/json"
	"fmt"
	"log"
	"xim/dispatcher/msgchan"
	"xim/logic"
)

func genQueueMsgTransformer() msgchan.MsgChannelTransformer {
	idGen := NewIDGenerator()
	return func(m interface{}) interface{} {
		qm := m.(*queueMsg)
		id := idGen.ID()
		qm.id <- id
		return &chanMsg{id, qm.channel, qm.user, qm.msgType, qm.msg}
	}
}

type msgChanTransformer struct {
	count  uint
	lastID string
}

func (t *msgChanTransformer) transform(m interface{}) interface{} {
	cm := m.(*chanMsg)
	// save to db.
	t.count++
	lastID := t.lastID
	t.lastID = cm.id
	log.Println("channel:", cm)
	return &toDispatchMsg{
		channel: cm.channel,
		user:    cm.user,
		id:      cm.id,
		lastID:  lastID,
		msgType: cm.msgType,
		msg:     cm.msg,
	}
}

func dispatchMsg(m interface{}) error {
	dm := m.(*toDispatchMsg)
	doDispatchMsg(dm.channel, dm.user, dm.msgType, dm.id, dm.lastID, dm.msg)
	return nil
}

func newDispatcherMsgChan(name string) *msgchan.MsgChannel {
	c := msgchan.NewMsgChannel(fmt.Sprintf("%s.channel", name), 100,
		new(msgChanTransformer).transform,
		msgchan.NewMsgChannelHandlerDownStream(fmt.Sprintf("%s.dispatcher", name), dispatchMsg))

	return msgchan.NewMsgChannel(fmt.Sprintf("%s.queue", name), 10, genQueueMsgTransformer(), c)
}

type queueMsg struct {
	user    logic.UserLocation
	channel string
	msgType string
	msg     json.RawMessage
	id      chan string
}

type chanMsg struct {
	id      string
	channel string
	user    logic.UserLocation
	msgType string
	msg     json.RawMessage
}

type toDispatchMsg struct {
	channel string
	user    logic.UserLocation
	id      string
	lastID  string
	msgType string
	msg     json.RawMessage
}

func (qm *queueMsg) String() string {
	return fmt.Sprintf("%s: %s", qm.user, string(qm.msg))
}

func (cm *chanMsg) String() string {
	return fmt.Sprintf("%s: %s[%s]", cm.user, string(cm.msg), cm.id)
}

func (cm *toDispatchMsg) String() string {
	return fmt.Sprintf("%s: %s[%s<-%s]", cm.user, string(cm.msg), cm.lastID, cm.id)
}
