package mid

import (
	"fmt"
	"math"
	"time"
	"xim/xchat/logic/db"
	pubtypes "xim/xchat/logic/pub/types"
	"xim/xchat/logic/service/types"
)

func handleMsg(ms <-chan interface{}) {
	for m := range ms {
		switch msg := m.(type) {
		case *pubtypes.Message:
			l.Info("push msg: %+v", msg)
			go push(msg)
		}
	}
}

type xsess struct {
	p      *PushState
	lastID uint64
	task   chan []*Message
}

func push(msg *pubtypes.Message) {
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

func pushSessMsg(x *Session, msg *pubtypes.Message) {
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

func pushSessesMsgs(xsesses []*xsess, minLastID uint64, msg *pubtypes.Message, include bool) {
	var msgs []pubtypes.Message
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
		go pushChatUserMsgs(p.pushing, p.s.ID, p.taskChan)
	default:
		// FIXME: BBBBBB => AAAAAA
		// run a goroutine to go pushChatUserMsgs periodly.
	}
}

func pushChatUserMsgs(pushing chan struct{}, id SessionID, tasks <-chan chan []*Message) {
	var accMsgs []*Message
	for {
		select {
		case task := <-tasks:
			msgs, ok := <-task
			if ok {
				accMsgs = append(accMsgs, msgs...)
				if len(accMsgs) > 10 {
					err := xchat.Publish(fmt.Sprintf(URIXChatUserMsg, id), []interface{}{accMsgs}, emptyKwargs)
					if err != nil {
						l.Warning("publish msg error:", err)
					}
					accMsgs = []*Message{}
				}
			}
		case <-time.After(12 * time.Millisecond):
			if len(accMsgs) > 0 {
				err := xchat.Publish(fmt.Sprintf(URIXChatUserMsg, id), []interface{}{accMsgs}, emptyKwargs)
				if err != nil {
					l.Warning("publish msg error:", err)
				}
				accMsgs = []*Message{}
			}
			select {
			case task := <-tasks:
				msgs, ok := <-task
				if ok {
					accMsgs = append(accMsgs, msgs...)
				}
			case <-time.After(3 * time.Second):
				pushing <- struct{}{}
				// FIXME: AAAAAA => BBBBBB
				return
			}
		}
	}
}
