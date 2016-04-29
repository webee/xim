package rpcservice

import (
	"fmt"
	"log"
	"xim/broker/proto"
	"xim/broker/userds"
	"xim/dispatcher/broker"
	"xim/dispatcher/msgchan"
)

func genQueueMsgTransformer() msgchan.MsgChannelTransformer {
	idGen := NewIDGenerator()
	return func(m interface{}) interface{} {
		qm := m.(*queueMsg)
		id := idGen.ID()
		qm.id <- id
		return &chanMsg{id, qm.channel, qm.user, qm.msg}
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
		msg:     cm.msg,
	}
}

func dispatchMsg(m interface{}) error {
	dm := m.(*toDispatchMsg)
	doDispatchMsg(dm.channel, &dm.user, dm.id, dm.lastID, dm.msg)
	return nil
}

func newDispatcherMsgChan(name string) *msgchan.MsgChannel {
	c := msgchan.NewMsgChannel(fmt.Sprintf("%s.channel", name), 100,
		new(msgChanTransformer).transform,
		msgchan.NewMsgChannelHandlerDownStream(fmt.Sprintf("%s.dispatcher", name), dispatchMsg))

	return msgchan.NewMsgChannel(fmt.Sprintf("%s.queue", name), 10, genQueueMsgTransformer(), c)
}

type queueMsg struct {
	user    userds.UserLocation
	channel string
	msg     interface{}
	id      chan string
}

type chanMsg struct {
	id      string
	channel string
	user    userds.UserLocation
	msg     interface{}
}

type toDispatchMsg struct {
	channel string
	user    userds.UserLocation
	id      string
	lastID  string
	msg     interface{}
}

func (qm *queueMsg) String() string {
	return fmt.Sprintf("%s: %s", qm.user, qm.msg)
}

func (cm *chanMsg) String() string {
	return fmt.Sprintf("%s: %s[%s]", cm.user, cm.msg, cm.id)
}

func (cm *toDispatchMsg) String() string {
	return fmt.Sprintf("%s: %s[%s<-%s]", cm.user, cm.msg, cm.lastID, cm.id)
}

func pushMsg(m interface{}) error {
	pm := m.(*toPushMsg)
	return broker.PushMsg(pm.user, pm.msg)
}

func newUserMsgChan(name string) *msgchan.MsgChannel {
	return msgchan.NewMsgChannel(fmt.Sprintf("user.%s.msgchan", name), 10, nil,
		msgchan.NewMsgChannelHandlerDownStream(fmt.Sprintf("user.%s.pusher", name), pushMsg))
}

type toPushMsg struct {
	user userds.UserLocation
	msg  proto.ChannelMsg
}
