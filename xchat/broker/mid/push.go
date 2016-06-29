package mid

import (
	"sort"
	"sync"
	pubtypes "xim/xchat/logic/pub/types"
	"xim/xchat/logic/service/types"
)

var (
	// TODO use cache.
	pushStatesLock = sync.RWMutex{}
	pushStates     = make(map[uint64]*PushState, 128)
)

// ByMsgID implements sort.Interface for []*pubtypes.ChatMessage by msg id desc.
type ByMsgID []*pubtypes.ChatMessage

func (a ByMsgID) Len() int           { return len(a) }
func (a ByMsgID) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByMsgID) Less(i, j int) bool { return a[i].ID < a[j].ID }

// PushState is chat's push state.
type PushState struct {
	sync.RWMutex
	chatID   uint64
	chatType string
	// only accept msg which id > pendingMsgID.
	pendingMsgID uint64
	pushedMsgID  uint64
	pushing      chan struct{}
	replace      chan struct{}
	msgs         []*pubtypes.ChatMessage
}

// Pending pending push message.
func (p *PushState) Pending(msg *pubtypes.ChatMessage) bool {
	p.Lock()
	if p.pendingMsgID >= msg.ID {
		if msg.ID > p.pushedMsgID {
			// 缓存消息
			p.msgs = append(p.msgs, msg)
		}
		p.Unlock()
		return false
	}
	p.msgs = append(p.msgs, msg)

	if p.pushedMsgID == 0 {
		// 初始状态
		p.pushedMsgID = msg.ID - 1
	}

	p.pendingMsgID = msg.ID
	close(p.replace)
	p.replace = make(chan struct{})
	p.Unlock()

	select {
	case p.pushing <- NT:
		return true
	case <-p.replace:
		return false
	}
}

// FetchMsgs fetch to push messages.
func (p *PushState) FetchMsgs(msgID uint64) []*pubtypes.ChatMessage {
	p.RLock()
	msgs := []*pubtypes.ChatMessage{}
	for _, msg := range p.msgs {
		if msg.ID <= msgID {
			msgs = append(msgs, msg)
		}
	}
	defer p.RUnlock()

	if uint64(len(msgs)) < msgID-p.pushedMsgID {
		missedIDs := []uint64{}
		for i := p.pushedMsgID + 1; i <= msgID; i++ {
			missed := true
			for _, msg := range msgs {
				if msg.ID == i {
					missed = false
					break
				}
			}
			if missed {
				missedIDs = append(missedIDs, i)
			}
		}

		var missedMsgs []pubtypes.ChatMessage
		args := &types.FetchChatMessagesByIDsArgs{
			ChatID:   p.chatID,
			ChatType: p.chatType,
			MsgIDs:   missedIDs,
		}
		if err := xchatLogic.Call(types.RPCXChatFetchChatMessagesByIDs, args, &msgs); err != nil {
			l.Warning("fetch chat[%d] messages by ids error: %+v, %s", p.chatID, args, err.Error())
		}
		for _, msg := range missedMsgs {
			msgs = append(msgs, &msg)
		}
	}

	sort.Sort(ByMsgID(msgs))
	return msgs
}

// Done done pushing.
func (p *PushState) Done(msgID uint64) {
	p.Lock()
	defer p.Unlock()

	p.pushedMsgID = msgID

	newMsgs := p.msgs[:0]
	for _, msg := range p.msgs {
		if msg.ID > msgID {
			newMsgs = append(newMsgs, msg)
		}
	}
	p.msgs = newMsgs
	<-p.pushing
}

func newPushState(chatID uint64, chatType string) *PushState {
	return &PushState{
		chatID:   chatID,
		chatType: chatType,
		pushing:  make(chan struct{}, 1),
		replace:  make(chan struct{}),
		msgs:     []*pubtypes.ChatMessage{},
	}
}

// GetChatPushState returns chat's push state.
func GetChatPushState(chatID uint64, chatType string) *PushState {
	pushStatesLock.Lock()
	defer pushStatesLock.Unlock()

	p, ok := pushStates[chatID]
	if !ok {
		p = newPushState(chatID, chatType)
		pushStates[chatID] = p
	}
	return p
}

// RemoveChatPustState remove chat's push state.
func RemoveChatPustState(chatID uint64) {
	pushStatesLock.Lock()
	defer pushStatesLock.Unlock()

	delete(pushStates, chatID)
}
