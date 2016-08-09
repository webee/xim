package mid

import (
	"time"
	pubtypes "xim/xchat/logic/pub/types"
	"xim/xchat/logic/service/types"
)

func handleMsg(ms <-chan interface{}) {
	for xm := range ms {
		xmsg, ok := xm.(*pubtypes.XMessage)
		if !ok {
			continue
		}
		src := xmsg.Source

		switch msg := xmsg.Msg.(type) {
		case pubtypes.ChatMessage:
			if src != nil && src.InstanceID == instanceID {
				continue
			}
			go push(src, &msg)
		case pubtypes.ChatNotifyMessage:
			go pushNotify(&msg)
		case pubtypes.UserNotifyMessage:
			go pushUserNotify(&msg)
		case pubtypes.SetAreaLimitCmd:
			go setAreaLimit(&msg)
		}
	}
}

func setAreaLimit(cmd *pubtypes.SetAreaLimitCmd) {
	rooms.SetAreaLimit(cmd.Limit)
}

func getChatSessions(chatType string, chatID uint64, updated int64) (sessions []*Session) {
	if chatType == types.ChatTypeRoom {
		ids := rooms.ChatMembers(chatID)
		for _, id := range ids {
			x, ok := GetSession(id)
			if ok {
				sessions = append(sessions, x)
			}
		}
	} else {
		members := getChatMembers(chatID, updated)

		for _, member := range members {
			sessions = append(sessions, GetUserSessions(member.User)...)
		}
	}

	return
}

func pushUserNotify(msg *pubtypes.UserNotifyMessage) {
	sesses := GetUserSessions(msg.User)
	toPushMsgs := []StatelessMsg{NewUserNotifyMessageFromPubMsg(msg)}

	for _, s := range sesses {
		s.taskChan.NewStatelessTask() <- toPushMsgs
		tryPushing(s, s.taskChan)
	}
}

func pushNotify(msg *pubtypes.ChatNotifyMessage) {
	sesses := getChatSessions(msg.ChatType, msg.ChatID, msg.Updated)
	toPushMsgs := []StatelessMsg{NewNotifyMessageFromPubMsg(msg)}

	for _, s := range sesses {
		s.taskChan.NewStatelessTask() <- toPushMsgs
		tryPushing(s, s.taskChan)
	}
}

func push(src *pubtypes.MsgSource, msg *pubtypes.ChatMessage) {
	sessions := getChatSessions(msg.ChatType, msg.ChatID, msg.Updated)
	if len(sessions) == 0 {
		return
	}

	pushState := GetChatPushState(msg.ChatID, msg.ChatType)
	if !pushState.Pending(msg) {
		return
	}
	defer pushState.Done(msg.ID)

	msgs := pushState.FetchMsgs(msg.ID)
	toPushMsgs := []*Message{}
	for _, msg := range msgs {
		toPushMsgs = append(toPushMsgs, NewMessageFromPubMsg(msg))
	}

	for _, s := range sessions {
		// NOTE: 在低速情况下保证不push自己发送的消息
		if src != nil && src.InstanceID == uint64(instanceID) && src.SessionID == uint64(s.ID) {
			if len(toPushMsgs) > 1 {
				s.taskChan.NewTask() <- toPushMsgs[:len(toPushMsgs)-1]
			}
		} else {
			s.taskChan.NewTask() <- toPushMsgs
		}
		tryPushing(s, s.taskChan)
	}
}

func tryPushing(s *Session, t *TaskChan) {
	select {
	case <-t.pushing:
		// start a goroutine.
		go pushChatUserMsgs(s, t)
	default:
	}
}

func pushChatUserMsgs(s *Session, t *TaskChan) {
	// mutex
	<-t.pushingMutex

	xpushChatUserMsgs(s, t, false)
}

func xpushChatUserMsgs(s *Session, t *TaskChan, clear bool) {
	pushing := t.pushing
	tasks := t.tasks
	statelessTasks := t.statelessTasks
	rawTasks := t.rawTasks

	var accMsgs []*Message
	var accStatelessMsgs []StatelessMsg
	for {
		select {
		case task := <-statelessTasks:
			msgs, ok := <-task
			if !ok || len(msgs) == 0 {
				continue
			}
			accStatelessMsgs = append(accStatelessMsgs, msgs...)
			if len(accStatelessMsgs) > 32 {
				doPushStatelessMsgs(s.msgTopic, accStatelessMsgs)
				accStatelessMsgs = []StatelessMsg{}
			}
		case task := <-tasks:
			msgs, ok := <-task
			if !ok || len(msgs) == 0 {
				continue
			}
			accMsgs = append(accMsgs, msgs...)
			if len(accMsgs) > 32 {
				doPush(s.msgTopic, types.MsgKindChat, accMsgs)
				accMsgs = []*Message{}
			}
		case task := <-rawTasks:
			msg, ok := <-task
			if !ok {
				continue
			}
			doPush(s.msgTopic, msg.Kind, msg.Msgs)
		case <-time.After(18 * time.Millisecond):
			if len(accStatelessMsgs) > 0 {
				if len(accStatelessMsgs) == 1 {
					doPush(s.msgTopic, accStatelessMsgs[0].Kind(), accStatelessMsgs)
				} else {
					doPushStatelessMsgs(s.msgTopic, accStatelessMsgs)
				}
				accStatelessMsgs = []StatelessMsg{}
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
			case task := <-statelessTasks:
				msgs, ok := <-task
				if ok {
					accStatelessMsgs = append(accStatelessMsgs, msgs...)
				}
			case task := <-tasks:
				msgs, ok := <-task
				if ok {
					accMsgs = append(accMsgs, msgs...)
				}
			case task := <-rawTasks:
				msg, ok := <-task
				if !ok {
					continue
				}
				doPush(s.msgTopic, msg.Kind, msg.Msgs)
			// TODO 根据消息发送状况确定等待时间
			case <-time.After(3 * time.Second):
				// 让下一位进入
				pushing <- NT

				// 清除剩余任务
				xpushChatUserMsgs(s, t, true)
				t.pushingMutex <- NT
				return
			}
		}
	}
}

func doPushStatelessMsgs(topic string, msgs []StatelessMsg) {
	// TODO: optimize.
	kindMsgs := make(map[string][]StatelessMsg)
	for _, msg := range msgs {
		kindMsgs[msg.Kind()] = append(kindMsgs[msg.Kind()], msg)
	}
	for kind, msgs := range kindMsgs {
		doPush(topic, kind, msgs)
	}
}

func doPush(topic string, kind string, payload interface{}) {
	err := xchat.Publish(topic, []interface{}{kind, payload}, emptyKwargs)
	if err != nil {
		l.Warning("publish msg error:", err)
	}
}

func pushRawMsg(user string, msg *RawMessage) {
	sesses := GetUserSessions(user)

	for _, s := range sesses {
		s.taskChan.NewRawTask() <- msg
		tryPushing(s, s.taskChan)
	}
}
