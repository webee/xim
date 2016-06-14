package mid

import (
	"math"
	"time"
	pubtypes "xim/xchat/logic/pub/types"
	"xim/xchat/logic/service/types"
)

func handleMsg(ms <-chan interface{}) {
	for m := range ms {
		switch msg := m.(type) {
		case pubtypes.ChatMessage:
			go push(&msg)
		case pubtypes.ChatNotifyMessage:
			go pushNotify(&msg)
		}
	}
}

func pushNotify(msg *pubtypes.ChatNotifyMessage) {
	sessions := getChatSessions(msg.IsRoomChat, msg.ChatID, msg.Updated)

	for _, x := range sessions {
		if x.User == msg.User {
			// 不需要发送给消息发送者
			continue
		}

		p, task := x.GetNotifyPushState(msg.ChatID)
		toPush := NewNotifyMessageFromDBMsg(msg)
		task <- []*NotifyMessage{toPush}
		tryPushing(p)
	}
}

type xsess struct {
	p      *PushState
	lastID uint64
	task   chan []*Message
}

func getChatSessions(isRoomChat bool, chatID uint64, updated int64) (sessions []*Session) {
	if !isRoomChat {
		members := getChatMembers(chatID, updated)

		for _, member := range members {
			ss := GetUserSessions(member.User)
			for _, x := range ss {
				sessions = append(sessions, x)
			}
		}
	} else {
		ids := rooms.ChatMembers(chatID)
		for _, id := range ids {
			x, ok := GetSession(id)
			if ok {
				sessions = append(sessions, x)
			}
		}
	}

	return
}

func push(msg *pubtypes.ChatMessage) {
	sessions := getChatSessions(msg.IsRoomChat, msg.ChatID, msg.Updated)

	minLastID := uint64(math.MaxUint64)
	xsesses := []*xsess{}

	// TODO: one chat per PushState, not one session. 使用全局发送状态，各session维护自己的任务队列
	for _, x := range sessions {
		p, task, lastID, ok := x.GetPushState(msg)
		if !ok {
			// already send.
			continue
		}
		if lastID < minLastID {
			minLastID = lastID
		}
		xsesses = append(xsesses, &xsess{p, lastID, task})
	}
	if len(xsesses) < 1 {
		return
	}

	pushSessesMsgs(xsesses, minLastID, msg, true)
}

func pushSessMsg(p *PushState, msg *pubtypes.ChatMessage) {
	task, lastID, ok := p.getTask(msg.ID, false)
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
			LID:    minLastID,
			RID:    msg.ID,
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
	}
}

func pushChatUserMsgs(p *PushState) {
	// mutex
	<-p.pushingMutex

	xpushChatUserMsgs(p, false)
}

func xpushChatUserMsgs(p *PushState, clear bool) {
	pushing := p.pushing
	s := p.s
	tasks := p.taskChan
	notifyTasks := p.notifyTaskChan

	var accMsgs []*Message
	var accNotifyMsgs []*NotifyMessage
	for {
		select {
		case task := <-notifyTasks:
			msgs, ok := <-task
			if !ok || len(msgs) == 0 {
				continue
			}
			accNotifyMsgs = append(accNotifyMsgs, msgs...)
			if len(accNotifyMsgs) > 20 {
				doPush(s.msgTopic, types.MsgKindChatNotify, accNotifyMsgs)
				accNotifyMsgs = []*NotifyMessage{}
			}
		case task := <-tasks:
			msgs, ok := <-task
			if !ok || len(msgs) == 0 {
				continue
			}
			accMsgs = append(accMsgs, msgs...)
			if len(accMsgs) > 20 {
				doPush(s.msgTopic, types.MsgKindChat, accMsgs)
				accMsgs = []*Message{}
			}
		case <-time.After(15 * time.Millisecond):
			if len(accNotifyMsgs) > 0 {
				doPush(s.msgTopic, types.MsgKindChatNotify, accNotifyMsgs)
				accNotifyMsgs = []*NotifyMessage{}
			}
			if len(accMsgs) > 0 {
				doPush(s.msgTopic, types.MsgKindChat, accMsgs)
				accMsgs = []*Message{}
			}

			if clear {
				// 完成清除剩余任务
				return
			}

			select {
			case task := <-notifyTasks:
				msgs, ok := <-task
				if ok {
					accNotifyMsgs = append(accNotifyMsgs, msgs...)
				}
			case task := <-tasks:
				msgs, ok := <-task
				if ok {
					accMsgs = append(accMsgs, msgs...)
				}
			// TODO 连接消息发送状况确定等待时间
			case <-time.After(3 * time.Second):
				pushing <- struct{}{}

				// 清除剩余任务
				xpushChatUserMsgs(p, true)
				p.pushingMutex <- struct{}{}
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
