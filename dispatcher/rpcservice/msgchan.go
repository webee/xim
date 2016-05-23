package rpcservice

import (
	"fmt"
	"log"
	"time"
	"xim/broker/proto"
	"xim/broker/userds"
	"xim/dispatcher/broker"
	"xim/dispatcher/db"
	"xim/dispatcher/msgchan"
)

func genQueueMsgTransformer(channel string) msgchan.MsgChannelTransformer {
	msgStore := db.GetMsgStore()
	defer msgStore.Close()

	idGen := NewIDGenerator()
	// FIXME: handle error.
	lastID, _ := msgStore.LastID(channel)
	idGen.SetID(lastID)
	return func(m interface{}) interface{} {
		msgStore := db.GetMsgStore()
		defer msgStore.Close()

		qm := m.(*queueMsg)
		id := idGen.ID()
		ts := time.Now().Unix()
		// NOTE: save to db.
		if err := msgStore.AddChannelMsg(qm.channel, id, ts, qm.user.User, qm.msg); err != nil {
			close(qm.id)
			close(qm.ts)
			return nil
		}

		qm.id <- id
		qm.ts <- ts
		return &chanMsg{id, qm.channel, qm.user, qm.msg, ts}
	}
}

type msgChanTransformer struct {
	count uint
}

func (t *msgChanTransformer) transform(m interface{}) interface{} {
	if m == nil {
		return nil
	}

	cm := m.(*chanMsg)
	t.count++
	log.Println("channel:", cm)
	return &toDispatchMsg{
		channel: cm.channel,
		user:    cm.user,
		id:      cm.id,
		msg:     cm.msg,
		ts:      cm.ts,
	}
}

func dispatchMsg(m interface{}) error {
	if m == nil {
		return nil
	}

	dm := m.(*toDispatchMsg)
	doDispatchMsg(dm.channel, &dm.user, dm.id, "", dm.msg, dm.ts)
	return nil
}

func newDispatcherMsgChan(channel string) *msgchan.MsgChannel {
	c := msgchan.NewMsgChannel(fmt.Sprintf("%s.channel", channel), 100,
		new(msgChanTransformer).transform,
		msgchan.NewMsgChannelHandlerDownStream(fmt.Sprintf("%s.dispatcher", channel), dispatchMsg))

	return msgchan.NewMsgChannel(fmt.Sprintf("%s.queue", channel), 10, genQueueMsgTransformer(channel), c)
}

type queueMsg struct {
	user    userds.UserLocation
	channel string
	msg     interface{}
	id      chan int
	ts      chan int64
}

type chanMsg struct {
	id      int
	channel string
	user    userds.UserLocation
	msg     interface{}
	ts      int64
}

type toDispatchMsg struct {
	channel string
	user    userds.UserLocation
	id      int
	msg     interface{}
	ts      int64
}

func (qm *queueMsg) String() string {
	return fmt.Sprintf("%s: %s", qm.user, qm.msg)
}

func (cm *chanMsg) String() string {
	return fmt.Sprintf("%s: %s[%d]", cm.user, cm.msg, cm.id)
}

func (cm *toDispatchMsg) String() string {
	return fmt.Sprintf("%s: %s[%d]", cm.user, cm.msg, cm.id)
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
	msg  *proto.Push
}
