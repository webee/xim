package mid

import (
	"math"
	"time"
	"xim/xchat/logic/db"
	pubtypes "xim/xchat/logic/pub/types"
	"xim/xchat/logic/service/types"
)

func handleMsg(ms <-chan interface{}) {
	for m := range ms {
		switch msg := m.(type) {
		case pubtypes.ChatMessage:
			l.Info("push msg: %+v", msg)
			go push(&msg)
		case pubtypes.ChatNotifyMessage:
			l.Info("push notify msg: %+v", msg)
			go pushNotify(&msg)
		}
	}
}

type xsess struct {
	p      *PushState
	lastID uint64
	task   chan []*Message
}

func push(msg *pubtypes.ChatMessage) {
	var members []db.Member
	// TODO: timeout cache rpc call.
	// call every 2 seconds.
	if err := xchatLogic.Call(types.RPCXChatFetchChatMembers, msg.ChatID, &members); err != nil {
		l.Warning("fetch chat[%d] members error: %s", msg.ChatID, err)
		return
	}
	minLastID := uint64(math.MaxUint64)
	xsesses := []*xsess{}

	// TODO: one chat per PushState, not one session.
	for _, member := range members {
		ss := GetUserSessions(member.User)
		for _, x := range ss {
			p, task, lastID, ok := x.GetPushState(msg.ChatID, msg.ID)
			if !ok {
				// already send.
				continue
			}
			if lastID < minLastID {
				minLastID = lastID
			}
			xsesses = append(xsesses, &xsess{p, lastID, task})
		}
	}
	if len(xsesses) < 1 {
		return
	}

	pushSessesMsgs(xsesses, minLastID, msg, true)
}

func pushSessMsg(x *Session, msg *pubtypes.ChatMessage) {
	p, task, lastID, ok := x.GetPushState(msg.ChatID, msg.ID)
	if !ok {
		return
	}
	if lastID+1 == msg.ID {
		// already send.
		close(task)
		tryPushing(p)
		return
	}

	xsesses := []*xsess{&xsess{p, lastID, task}}
	pushSessesMsgs(xsesses, lastID, msg, false)
}

func pushSessesMsgs(xsesses []*xsess, minLastID uint64, msg *pubtypes.ChatMessage, include bool) {
	var msgs []pubtypes.ChatMessage
	if minLastID+1 < msg.ID {
		// fetch late messages.
		args := &types.FetchChatMessagesArgs{
			ChatID: msg.ChatID,
			SID:    minLastID,
			EID:    msg.ID,
		}
		if err := xchatLogic.Call(types.RPCXChatFetchChatMessages, args, &msgs); err != nil {
			l.Warning("fetch chat[%d] messages error: %s", msg.ChatID, err)
		}
	}
	if include {
		msgs = append(msgs, *msg)
	}

	toPushMsgs := []*Message{}
	for _, msg := range msgs {
		toPushMsgs = append(toPushMsgs, NewMessageFromDBMsg(&msg))
	}

	for _, xs := range xsesses {
		var toPush []*Message
		if xs.lastID+1 == msg.ID || xs.lastID == 0 {
			toPush = toPushMsgs[len(toPushMsgs)-1:]
		} else {
			for _, m := range toPushMsgs {
				if m.ID > xs.lastID && m.ID <= msg.ID {
					toPush = append(toPush, m)
				}
			}
		}

		xs.task <- toPush
		tryPushing(xs.p)
	}
}

func tryPushing(p *PushState) {
	select {
	case <-p.pushing:
		// start a goroutine.
		go pushChatUserMsgs(p)
	default:
		// FIXME: BBBBBB => AAAAAA
		// run a goroutine to go pushChatUserMsgs periodly.
	}
}

func pushChatUserMsgs(p *PushState) {
	pushing := p.pushing
	s := p.s
	tasks := p.taskChan
	notifyTasks := p.notifyTaskChan

	var accMsgs []*Message
	var accNotifyMsgs []*NotifyMessage
	for {
		select {
		case task := <-tasks:
			msgs, ok := <-task
			if !ok || len(msgs) == 0 {
				continue
			}
			accMsgs = append(accMsgs, msgs...)
			if len(accMsgs) > 10 {
				doPush(s.msgTopic, types.MsgKindChat, accMsgs)
				accMsgs = []*Message{}
			}
		case task := <-notifyTasks:
			msgs, ok := <-task
			if !ok || len(msgs) == 0 {
				continue
			}
			accNotifyMsgs = append(accNotifyMsgs, msgs...)
			if len(accNotifyMsgs) > 10 {
				doPush(s.msgTopic, types.MsgKindChatNotify, accNotifyMsgs)
				accNotifyMsgs = []*NotifyMessage{}
			}
		case <-time.After(12 * time.Millisecond):
			if len(accMsgs) > 0 {
				doPush(s.msgTopic, types.MsgKindChat, accMsgs)
				accMsgs = []*Message{}
			}
			if len(accNotifyMsgs) > 0 {
				doPush(s.msgTopic, types.MsgKindChatNotify, accNotifyMsgs)
				accNotifyMsgs = []*NotifyMessage{}
			}

			select {
			case task := <-tasks:
				msgs, ok := <-task
				if ok {
					accMsgs = append(accMsgs, msgs...)
				}
			case task := <-notifyTasks:
				msgs, ok := <-task
				if ok {
					accNotifyMsgs = append(accNotifyMsgs, msgs...)
				}
			case <-time.After(3 * time.Second):
				pushing <- struct{}{}
				// FIXME: AAAAAA => BBBBBB
				return
			}
		}
	}
}

func doPush(topic string, kind string, payload interface{}) {
	err := xchat.Publish(topic, []interface{}{kind, payload}, emptyKwargs)
	if err != nil {
		l.Warning("publish msg error:", err)
	}
}

func pushNotify(msg *pubtypes.ChatNotifyMessage) {
	var members []db.Member
	// TODO: timeout cache rpc call.
	// call every 2 seconds.
	if err := xchatLogic.Call(types.RPCXChatFetchChatMembers, msg.ChatID, &members); err != nil {
		l.Warning("fetch chat[%d] members error: %s", msg.ChatID, err)
		return
	}

	for _, member := range members {
		ss := GetUserSessions(member.User)
		for _, x := range ss {
			p, task := x.GetNotifyPushState(msg.ChatID)
			toPush := NewNotifyMessageFromDBMsg(msg)
			task <- []*NotifyMessage{toPush}
			tryPushing(p)
		}
	}
}
