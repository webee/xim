package mid

import (
	"fmt"
	"math"
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
	seq    uint64
}

func push(msg *pubtypes.Message) {
	var members []db.Member
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
			p, seq, lastID, ok := x.GetSetPushID(msg.ChatID, msg.ID)
			if !ok {
				// already send.
				continue
			}
			if lastID > 0 && lastID < minLastID {
				minLastID = lastID
			}
			xsesses = append(xsesses, &xsess{p, lastID, seq})
		}
	}
	if len(xsesses) < 1 {
		return
	}

	pushSessesMsgs(xsesses, minLastID, msg, true, true)
}

func pushSessMsg(x *Session, msg *pubtypes.Message) {
	p, seq, lastID, ok := x.GetSetPushID(msg.ChatID, msg.ID)
	if !ok {
		return
	}
	if lastID == 0 || lastID+1 == msg.ID {
		// already send.
		p.DonePushing(seq)
		return
	}

	xsesses := []*xsess{&xsess{p, lastID, seq}}
	pushSessesMsgs(xsesses, lastID, msg, false, false)
}

func pushSessesMsgs(xsesses []*xsess, minLastID uint64, msg *pubtypes.Message, include bool, async bool) {
	l.Info("minLastID: %d, count: %d", minLastID, len(xsesses))
	var msgs []pubtypes.Message
	if minLastID != math.MaxUint64 && minLastID < msg.ID-1 {
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

		// TODO: check and optimize.
		//
		if async {
			go doPush(xs.seq, xs.p, toPush)
		} else {
			doPush(xs.seq, xs.p, toPush)
		}
	}
}

func doPush(seq uint64, p *PushState, msgs []*Message) {
	l.Info("push sess: %d, %d:%d, pushed: %d", p.s.ID, seq, p.curSeq, p.pushMsgID)
	if ok := p.Pushing(seq); !ok {
		return
	}
	defer p.DonePushing(seq)

	err := xchat.Publish(fmt.Sprintf(URIXChatUserMsg, p.s.ID), []interface{}{msgs}, emptyKwargs)
	if err != nil {
		l.Warning("publish msg error:", err)
	}
}
