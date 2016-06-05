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
	s      *Session
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

	for _, member := range members {
		ss := GetUserSessions(member.User)
		for _, x := range ss {
			seq, lastID, ok := x.GetSetPushID(msg.MsgID)
			if !ok {
				// already send.
				continue
			}
			if lastID > 0 && lastID < minLastID {
				minLastID = lastID
			}
			xsesses = append(xsesses, &xsess{x, lastID, seq})
		}
	}
	if len(xsesses) < 1 {
		return
	}

	pushSessesMsgs(xsesses, minLastID, msg)
}

func pushSessMsg(x *Session, msg *pubtypes.Message) {
	seq, lastID, ok := x.GetSetPushID(msg.MsgID)
	if !ok {
		return
	}
	if lastID == 0 || lastID+1 == msg.MsgID {
		// already send.
		x.DoneSending(seq)
		return
	}

	xsesses := []*xsess{&xsess{x, lastID, seq}}
	pushSessesMsgs(xsesses, lastID, msg)
}

func pushSessesMsgs(xsesses []*xsess, minLastID uint64, msg *pubtypes.Message) {
	l.Info("minLastID: %d, count: %d", minLastID, len(xsesses))
	var msgs []pubtypes.Message
	if minLastID != math.MaxUint64 && minLastID < msg.MsgID-1 {
		// fetch late messages.
		args := &types.FetchChatMessagesArgs{
			ChatID: msg.ChatID,
			SID:    minLastID,
			EID:    msg.MsgID,
		}
		if err := xchatLogic.Call(types.RPCXChatFetchChatMessages, args, &msgs); err != nil {
			l.Warning("fetch chat[%d] messages error: %s", msg.ChatID, err)
		}
	}
	msgs = append(msgs, *msg)

	toPushMsgs := []*Message{}
	for _, msg := range msgs {
		toPushMsgs = append(toPushMsgs, NewMessageFromDBMsg(&msg))
	}

	for _, xs := range xsesses {
		var toPush []*Message
		if xs.lastID+1 == msg.MsgID || xs.lastID == 0 {
			toPush = toPushMsgs[len(toPushMsgs)-1:]
		} else {
			for _, m := range toPushMsgs {
				if m.MsgID > xs.lastID && m.MsgID <= msg.MsgID {
					toPush = append(toPush, m)
				}
			}
		}

		// TODO: check
		//
		go func(seq uint64, s *Session) {
			l.Info("push sess: %d, %d:%d, pushed: %d", s.ID, seq, s.curSeq, s.pushMsgID)
			if ok := s.Sending(seq); !ok {
				return
			}
			defer s.DoneSending(seq)

			err := xchat.Publish(fmt.Sprintf(URIXChatUserMsg, s.ID), []interface{}{toPush}, emptyKwargs)
			if err != nil {
				l.Warning("publish msg error:", err)
			}
		}(xs.seq, xs.s)
	}
}
